import { Mutex } from 'async-mutex';

class RWLock {
  private readers = 0;
  private readerCountMutex = new Mutex();
  private writerMutex = new Mutex();
  private writerRelease: (() => void) | null = null;

  async acquireRead(): Promise<() => Promise<void>> {
    const releaseCount = await this.readerCountMutex.acquire();
    this.readers++;
    if (this.readers === 1) {
      const rel = await this.writerMutex.acquire();
      this.writerRelease = rel;
    }
    releaseCount();
    return async () => {
      const releaseCount2 = await this.readerCountMutex.acquire();
      this.readers--;
      if (this.readers === 0 && this.writerRelease) {
        this.writerRelease();
        this.writerRelease = null;
      }
      releaseCount2();
    };
  }

  async acquireWrite(): Promise<() => void> {
    return await this.writerMutex.acquire();
  }
}

export class FileLockManager {
  private locks = new Map<string, RWLock>();

  private getLock(filePath: string): RWLock {
    let lock = this.locks.get(filePath);
    if (!lock) {
      lock = new RWLock();
      this.locks.set(filePath, lock);
    }
    return lock;
  }

  async withRead<T>(filePath: string, fn: () => Promise<T>): Promise<T> {
    const lock = this.getLock(filePath);
    const release = await lock.acquireRead();
    try {
      return await fn();
    } finally {
      await release();
    }
  }

  async withWrite<T>(filePath: string, fn: () => Promise<T>): Promise<T> {
    const lock = this.getLock(filePath);
    const release = await lock.acquireWrite();
    try {
      return await fn();
    } finally {
      release();
    }
  }
}
