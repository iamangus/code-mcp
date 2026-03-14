import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { WorktreeJail, ToolError } from '../worktree.js';
import fs from 'fs';
import os from 'os';
import path from 'path';

let tmpDir: string;

beforeAll(() => {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'worktree-test-'));
});

afterAll(() => {
  fs.rmSync(tmpDir, { recursive: true, force: true });
});

describe('WorktreeJail', () => {
  it('resolves relative paths correctly', () => {
    const jail = new WorktreeJail(tmpDir);
    const resolved = jail.resolve('foo/bar.txt');
    expect(resolved).toBe(path.join(tmpDir, 'foo', 'bar.txt'));
  });

  it('throws ToolError for paths escaping root', () => {
    const jail = new WorktreeJail(tmpDir);
    expect(() => jail.resolve('../../../etc/passwd')).toThrow(ToolError);
    expect(() => jail.resolve('../../../etc/passwd')).toThrow(/escapes the worktree root/);
  });

  it('throws ToolError for null byte paths', () => {
    const jail = new WorktreeJail(tmpDir);
    expect(() => jail.resolve('foo\0bar')).toThrow(ToolError);
    expect(() => jail.resolve('foo\0bar')).toThrow(/null bytes/);
  });

  it('throws Error for non-existent root', () => {
    expect(() => new WorktreeJail('/nonexistent/path/xyz123')).toThrow(/does not exist/);
  });

  it('accepts root path itself', () => {
    const jail = new WorktreeJail(tmpDir);
    const resolved = jail.resolve('.');
    expect(resolved).toBe(tmpDir);
  });
});
