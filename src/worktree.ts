import path from 'path';
import fs from 'fs';

export class ToolError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ToolError';
  }
}

export class WorktreeJail {
  readonly root: string;

  constructor(rootPath: string) {
    this.root = path.resolve(rootPath);
    if (!fs.existsSync(this.root)) {
      throw new Error(`Worktree root does not exist: ${this.root}`);
    }
  }

  resolve(relativePath: string): string {
    if (relativePath.includes('\0')) {
      throw new ToolError('Tool Error: Path contains null bytes.');
    }
    const normalized = path.normalize(relativePath);
    const absolute = path.resolve(this.root, normalized);
    if (!absolute.startsWith(this.root + path.sep) && absolute !== this.root) {
      throw new ToolError(
        `Tool Error: Path "${relativePath}" escapes the worktree root. All paths must be relative to the worktree root.`
      );
    }
    return absolute;
  }
}
