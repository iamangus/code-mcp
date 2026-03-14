import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { z } from 'zod';
import { WorktreeJail, ToolError } from './worktree.js';
import { FileLockManager } from './locks.js';
import {
  readFile,
  readLines,
  createFile,
  listDirectory,
  grepSearch,
  searchAndReplace,
} from './tools/filesystem.js';
import { executeTerminalCommand, getGitDiff } from './tools/cli.js';

function errorResponse(err: unknown) {
  const message = err instanceof Error ? err.message : String(err);
  return { content: [{ type: 'text' as const, text: message }], isError: true as const };
}

export function createServer(jail: WorktreeJail): McpServer {
  const locks = new FileLockManager();
  const server = new McpServer({
    name: 'code-mcp',
    version: '1.0.0',
  });

  server.tool(
    'read_file',
    'Read the full content of a file within the worktree',
    {
      path: z.string().describe('Relative path to the file within the worktree'),
    },
    async ({ path: filePath }) => {
      try {
        const content = await readFile(jail, locks, filePath);
        return { content: [{ type: 'text', text: content }] };
      } catch (err) {
        if (err instanceof ToolError) return errorResponse(err);
        throw err;
      }
    }
  );

  server.tool(
    'read_lines',
    'Read specific lines from a file within the worktree',
    {
      path: z.string().describe('Relative path to the file within the worktree'),
      start_line: z.number().int().min(1).describe('Start line number (1-indexed, inclusive)'),
      end_line: z.number().int().min(1).describe('End line number (1-indexed, inclusive)'),
    },
    async ({ path: filePath, start_line, end_line }) => {
      try {
        const content = await readLines(jail, locks, filePath, start_line, end_line);
        return { content: [{ type: 'text', text: content }] };
      } catch (err) {
        if (err instanceof ToolError) return errorResponse(err);
        throw err;
      }
    }
  );

  server.tool(
    'create_file',
    'Create a new file with content. Fails if the file already exists.',
    {
      path: z.string().describe('Relative path to the new file within the worktree'),
      content: z.string().describe('Content of the file'),
    },
    async ({ path: filePath, content }) => {
      try {
        const result = await createFile(jail, locks, filePath, content);
        return { content: [{ type: 'text', text: result }] };
      } catch (err) {
        if (err instanceof ToolError) return errorResponse(err);
        throw err;
      }
    }
  );

  server.tool(
    'list_directory',
    'List files and directories in a path within the worktree',
    {
      path: z.string().describe('Relative path to the directory within the worktree'),
      recursive: z.boolean().optional().describe('Whether to list recursively'),
    },
    async ({ path: dirPath, recursive }) => {
      try {
        const result = await listDirectory(jail, dirPath, recursive ?? false);
        return { content: [{ type: 'text', text: result }] };
      } catch (err) {
        if (err instanceof ToolError) return errorResponse(err);
        throw err;
      }
    }
  );

  server.tool(
    'grep_search',
    'Search for a pattern in files within the worktree',
    {
      pattern: z.string().describe('The search pattern (string or regex)'),
      directory: z.string().optional().describe('Restrict search to this directory (relative path)'),
      is_regex: z.boolean().optional().describe('Whether the pattern is a regex'),
    },
    async ({ pattern, directory, is_regex }) => {
      try {
        const result = await grepSearch(jail, pattern, directory, is_regex ?? false);
        return { content: [{ type: 'text', text: result }] };
      } catch (err) {
        if (err instanceof ToolError) return errorResponse(err);
        throw err;
      }
    }
  );

  server.tool(
    'search_and_replace',
    'Search for a block of text in a file and replace it. Tries exact match first, then fuzzy matching.',
    {
      path: z.string().describe('Relative path to the file within the worktree'),
      search_block: z.string().describe('The block of text to search for'),
      replace_block: z.string().describe('The block of text to replace with'),
    },
    async ({ path: filePath, search_block, replace_block }) => {
      try {
        const result = await searchAndReplace(jail, locks, filePath, search_block, replace_block);
        return { content: [{ type: 'text', text: result }] };
      } catch (err) {
        if (err instanceof ToolError) return errorResponse(err);
        throw err;
      }
    }
  );

  server.tool(
    'execute_terminal_command',
    'Execute a terminal command in the worktree root directory',
    {
      command: z.string().describe('The shell command to execute'),
    },
    async ({ command }) => {
      try {
        const result = await executeTerminalCommand(jail, command);
        const output = [
          `Exit code: ${result.exitCode}`,
          result.timedOut ? 'TIMED OUT after 120 seconds' : '',
          result.stdout ? `STDOUT:\n${result.stdout}` : '',
          result.stderr ? `STDERR:\n${result.stderr}` : '',
        ]
          .filter(Boolean)
          .join('\n');
        return { content: [{ type: 'text', text: output }] };
      } catch (err) {
        if (err instanceof ToolError) return errorResponse(err);
        throw err;
      }
    }
  );

  server.tool(
    'get_git_diff',
    'Get the git diff for the current worktree (git diff HEAD + untracked files)',
    {},
    async () => {
      try {
        const diff = await getGitDiff(jail);
        return { content: [{ type: 'text', text: diff || '(no changes)' }] };
      } catch (err) {
        if (err instanceof ToolError) return errorResponse(err);
        throw err;
      }
    }
  );

  return server;
}
