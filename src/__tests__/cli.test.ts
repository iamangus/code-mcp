import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { WorktreeJail } from '../worktree.js';
import { executeTerminalCommand, getGitDiff } from '../tools/cli.js';
import fs from 'fs';
import os from 'os';
import path from 'path';
import { execSync } from 'child_process';

let tmpDir: string;
let jail: WorktreeJail;

beforeAll(() => {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cli-test-'));
  jail = new WorktreeJail(tmpDir);
});

afterAll(() => {
  fs.rmSync(tmpDir, { recursive: true, force: true });
});

describe('execute_terminal_command', () => {
  it('runs echo and captures stdout', async () => {
    const result = await executeTerminalCommand(jail, 'echo "hello world"');
    expect(result.stdout.trim()).toBe('hello world');
    expect(result.exitCode).toBe(0);
    expect(result.timedOut).toBe(false);
  });

  it('captures exit code', async () => {
    const result = await executeTerminalCommand(jail, 'exit 42');
    expect(result.exitCode).toBe(42);
  });

  it('captures stderr', async () => {
    const result = await executeTerminalCommand(jail, 'echo "error message" >&2');
    expect(result.stderr.trim()).toBe('error message');
  });

  it('times out with very short timeout', async () => {
    const result = await executeTerminalCommand(jail, 'sleep 10', 100);
    expect(result.timedOut).toBe(true);
    expect(result.exitCode).toBe(-1);
  });
});

describe('get_git_diff', () => {
  it('returns diff string in a git repo', async () => {
    execSync('git init', { cwd: tmpDir });
    execSync('git config user.email "test@test.com"', { cwd: tmpDir });
    execSync('git config user.name "Test"', { cwd: tmpDir });
    fs.writeFileSync(path.join(tmpDir, 'test.txt'), 'initial content');
    execSync('git add .', { cwd: tmpDir });
    execSync('git commit -m "initial"', { cwd: tmpDir });
    fs.writeFileSync(path.join(tmpDir, 'test.txt'), 'modified content');

    const diff = await getGitDiff(jail);
    expect(typeof diff).toBe('string');
    expect(diff).toContain('modified content');
  });
});
