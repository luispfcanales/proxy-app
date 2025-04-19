module.exports = {
  apps: [{
    name: "proxy-app",
    script: "./main.exe",
    watch: false,
    instances: 1,
    exec_mode: "fork",
    env: {
      NODE_ENV: "production",
      PORT: 80
    }
  }]
}