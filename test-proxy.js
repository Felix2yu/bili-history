const { spawn } = require('child_process');
const http = require('http');

const proc = spawn('node', ['.output/server/index.mjs'], {
  cwd: '/workspace/frontend',
  env: { ...process.env, NITRO_PORT: '3099', NUXT_BACKEND_URL: 'http://httpbin.org' },
  stdio: ['pipe', 'pipe', 'pipe']
});

let output = '';
proc.stdout.on('data', d => output += d.toString());
proc.stderr.on('data', d => output += d.toString());

setTimeout(() => {
  http.get('http://127.0.0.1:3099/api/history/test', (res) => {
    let data = '';
    res.on('data', d => data += d);
    res.on('end', () => {
      console.log('Status:', res.statusCode);
      console.log('Body:', data.slice(0, 500));
      console.log('--- server output ---');
      console.log(output.slice(-1000));
      proc.kill();
    });
  }).on('error', e => {
    console.log('Request error:', e.message);
    console.log('--- server output ---');
    console.log(output.slice(-1000));
    proc.kill();
  });
}, 3000);
