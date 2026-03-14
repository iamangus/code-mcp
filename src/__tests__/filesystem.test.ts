import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { WorktreeJail, ToolError } from '../worktree.js';
import { FileLockManager } from '../locks.js';
import {
  readFile,
  readLines,
  createFile,
  listDirectory,
  grepSearch,
  searchAndReplace,
} from '../tools/filesystem.js';
import fs from 'fs';
import os from 'os';
import path from 'path';

let tmpDir: string;
let jail: WorktreeJail;
let locks: FileLockManager;

beforeAll(async () => {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'fs-test-'));
  jail = new WorktreeJail(tmpDir);
  locks = new FileLockManager();
});

afterAll(async () => {
  fs.rmSync(tmpDir, { recursive: true, force: true });
});

describe('read_file', () => {
  it('reads file content', async () => {
    const filePath = path.join(tmpDir, 'hello.txt');
    fs.writeFileSync(filePath, 'Hello, World!');
    const content = await readFile(jail, locks, 'hello.txt');
    expect(content).toBe('Hello, World!');
  });

  it('throws ToolError for non-existent file', async () => {
    await expect(readFile(jail, locks, 'nonexistent.txt')).rejects.toThrow(ToolError);
    await expect(readFile(jail, locks, 'nonexistent.txt')).rejects.toThrow(/not found/);
  });

  it('throws ToolError for file exceeding size limit', async () => {
    const bigPath = path.join(tmpDir, 'bigfile.txt');
    const buf = Buffer.alloc(1_048_577, 'a');
    fs.writeFileSync(bigPath, buf);
    await expect(readFile(jail, locks, 'bigfile.txt')).rejects.toThrow(ToolError);
    await expect(readFile(jail, locks, 'bigfile.txt')).rejects.toThrow(/exceeds the maximum size/);
  });
});

describe('read_lines', () => {
  it('reads correct lines', async () => {
    const filePath = path.join(tmpDir, 'multiline.txt');
    fs.writeFileSync(filePath, 'line1\nline2\nline3\nline4\nline5');
    const content = await readLines(jail, locks, 'multiline.txt', 2, 4);
    expect(content).toBe('line2\nline3\nline4');
  });

  it('throws ToolError for non-existent file', async () => {
    await expect(readLines(jail, locks, 'nope.txt', 1, 2)).rejects.toThrow(ToolError);
  });
});

describe('create_file', () => {
  it('creates a new file', async () => {
    const result = await createFile(jail, locks, 'newfile.txt', 'content here');
    expect(result).toContain('created successfully');
    const content = fs.readFileSync(path.join(tmpDir, 'newfile.txt'), 'utf-8');
    expect(content).toBe('content here');
  });

  it('creates parent directories as needed', async () => {
    const result = await createFile(jail, locks, 'subdir/deep/file.txt', 'nested content');
    expect(result).toContain('created successfully');
    const content = fs.readFileSync(path.join(tmpDir, 'subdir/deep/file.txt'), 'utf-8');
    expect(content).toBe('nested content');
  });

  it('throws ToolError if file already exists', async () => {
    await expect(createFile(jail, locks, 'newfile.txt', 'new content')).rejects.toThrow(ToolError);
    await expect(createFile(jail, locks, 'newfile.txt', 'new content')).rejects.toThrow(/already exists/);
  });
});

describe('list_directory', () => {
  beforeAll(async () => {
    const testDir = path.join(tmpDir, 'listtest');
    fs.mkdirSync(testDir);
    fs.writeFileSync(path.join(testDir, 'a.txt'), 'a');
    fs.writeFileSync(path.join(testDir, 'b.txt'), 'b');
    fs.mkdirSync(path.join(testDir, 'subdir'));
    fs.writeFileSync(path.join(testDir, 'subdir', 'c.txt'), 'c');
    fs.mkdirSync(path.join(testDir, 'node_modules'));
    fs.writeFileSync(path.join(testDir, 'node_modules', 'pkg.js'), 'pkg');
    fs.mkdirSync(path.join(testDir, '.git'));
    fs.writeFileSync(path.join(testDir, '.git', 'HEAD'), 'ref: HEAD');
  });

  it('lists files non-recursively', async () => {
    const result = await listDirectory(jail, 'listtest', false);
    expect(result).toContain('a.txt');
    expect(result).toContain('b.txt');
    expect(result).toContain('subdir/');
  });

  it('ignores node_modules and .git in non-recursive mode', async () => {
    const result = await listDirectory(jail, 'listtest', false);
    expect(result).not.toContain('node_modules');
    expect(result).not.toContain('.git');
  });

  it('lists files recursively', async () => {
    const result = await listDirectory(jail, 'listtest', true);
    expect(result).toContain('a.txt');
    expect(result).toContain('b.txt');
    expect(result).toContain('subdir/c.txt');
  });

  it('ignores node_modules and .git in recursive mode', async () => {
    const result = await listDirectory(jail, 'listtest', true);
    expect(result).not.toContain('node_modules');
    expect(result).not.toContain('.git');
  });

  it('throws ToolError for non-existent directory', async () => {
    await expect(listDirectory(jail, 'nonexistent_dir', false)).rejects.toThrow(ToolError);
  });
});

describe('grep_search', () => {
  beforeAll(async () => {
    const searchDir = path.join(tmpDir, 'searchtest');
    fs.mkdirSync(searchDir);
    fs.writeFileSync(path.join(searchDir, 'file1.txt'), 'hello world\nfoo bar\nbaz');
    fs.writeFileSync(path.join(searchDir, 'file2.txt'), 'world peace\nhello again');
  });

  it('finds matches', async () => {
    const result = await grepSearch(jail, 'hello', 'searchtest');
    expect(result).toContain('hello');
    expect(result).toContain('file1.txt');
    expect(result).toContain('file2.txt');
  });

  it('returns empty message for no matches', async () => {
    const result = await grepSearch(jail, 'zzznomatch999', 'searchtest');
    expect(result).toBe('No matches found.');
  });

  it('supports regex patterns', async () => {
    const result = await grepSearch(jail, 'hell[o]', 'searchtest', true);
    expect(result).toContain('hello');
  });
});

describe('search_and_replace', () => {
  it('replaces a unique match', async () => {
    const filePath = path.join(tmpDir, 'replace_test.txt');
    fs.writeFileSync(filePath, 'line1\nfoo bar\nline3');
    const result = await searchAndReplace(jail, locks, 'replace_test.txt', 'foo bar', 'replaced');
    expect(result).toContain('File updated successfully');
    const content = fs.readFileSync(filePath, 'utf-8');
    expect(content).toContain('replaced');
    expect(content).not.toContain('foo bar');
  });

  it('fails for multiple matches', async () => {
    const filePath = path.join(tmpDir, 'multi_match.txt');
    fs.writeFileSync(filePath, 'foo\nfoo\nbar');
    await expect(
      searchAndReplace(jail, locks, 'multi_match.txt', 'foo', 'baz')
    ).rejects.toThrow(ToolError);
    await expect(
      searchAndReplace(jail, locks, 'multi_match.txt', 'foo', 'baz')
    ).rejects.toThrow(/matches \d+ locations/);
  });

  it('uses fuzzy matching for whitespace differences', async () => {
    const filePath = path.join(tmpDir, 'fuzzy_test.txt');
    fs.writeFileSync(filePath, 'function foo() {\n    const x = 1;\n    return x;\n}');
    const result = await searchAndReplace(
      jail,
      locks,
      'fuzzy_test.txt',
      'function foo() {\n  const x = 1;\n  return x;\n}',
      'function foo() {\n  const x = 2;\n  return x;\n}'
    );
    expect(result).toContain('File updated successfully');
  });
});
