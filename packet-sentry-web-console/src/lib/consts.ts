export const LOCALSTORAGE = {
  ADMIN_UI_ACCESS_TOKEN: "adminUiAccessToken",
  ADMIN_UI_REFRESH_TOKEN: "adminUiRefreshToken",
  API_ACCESS_TOKEN: "apiAccessToken",
  API_REFRESH_TOKEN: "apiRefreshToken",
} as const;

export const INSTALLER_INSTRUCTIONS = {
  "Ubuntu x64": `sudo apt-get install libpcap0.8\nsudo dpkg -i <installer>`,
  "macOS x64": "sudo installer -pkg <installer> -target /",
  "Ubuntu ARM64": `sudo apt-get install libpcap0.8\nsudo dpkg -i <installer>`,
  "macOS ARM64": "sudo installer -pkg <installer> -target /",
  "Windows x64": `The packet-sentry-agent on Windows depends on npcap.\nDownload the installer here: https://npcap.com/#download\n
    Once installed, you can run the packet-sentry-agent MSI.`,
  Unknown: "",
};
