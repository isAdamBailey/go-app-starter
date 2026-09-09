module.exports = {
  apps: [
    {
      // TODO: rename to your app (also update scripts/forge-deploy.sh's `pm2 delete` line)
      name: 'app-web',
      script: '.output/server/index.mjs',
      cwd: __dirname,
      instances: 1,
      exec_mode: 'fork',
      env: {
        NODE_ENV: 'production',
        HOST: '127.0.0.1',
      },
    },
  ],
}
