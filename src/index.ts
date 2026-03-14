#!/usr/bin/env node
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { StreamableHTTPServerTransport } from '@modelcontextprotocol/sdk/server/streamableHttp.js';
import { createServer } from './server.js';
import { WorktreeJail } from './worktree.js';
import http from 'http';

async function main() {
  const args = process.argv.slice(2);
  const useHttp = args.includes('--http');

  const worktreeRoot =
    args.find((a) => !a.startsWith('--')) ||
    process.env.WORKTREE_ROOT ||
    process.cwd();

  const jail = new WorktreeJail(worktreeRoot);
  const server = createServer(jail);

  if (useHttp) {
    const port = parseInt(process.env.PORT ?? '3000', 10);
    const httpServer = http.createServer(async (req, res) => {
      const transport = new StreamableHTTPServerTransport({
        sessionIdGenerator: undefined,
      });
      await server.connect(transport);
      const body = await getRawBody(req);
      let parsedBody: unknown;
      try {
        parsedBody = body.length > 0 ? JSON.parse(body.toString('utf-8')) : undefined;
      } catch {
        parsedBody = undefined;
      }
      await transport.handleRequest(req, res, parsedBody);
    });
    httpServer.listen(port, () => {
      console.error(`code-mcp HTTP server listening on port ${port}`);
    });
  } else {
    const transport = new StdioServerTransport();
    await server.connect(transport);
  }
}

async function getRawBody(req: http.IncomingMessage): Promise<Buffer> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];
    req.on('data', (chunk: Buffer) => chunks.push(chunk));
    req.on('end', () => resolve(Buffer.concat(chunks)));
    req.on('error', reject);
  });
}

main().catch((err) => {
  console.error('Fatal error:', err);
  process.exit(1);
});
