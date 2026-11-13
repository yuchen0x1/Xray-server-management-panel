3X-UI for Windows (amd64)

Extract the entire archive and keep the bin directory beside x-ui.exe.
Run x-ui.exe from a terminal and use the panel address shown in its output.
The process must remain running for the panel to stay available.
Use x-ui.exe help to inspect the supported command-line options.

Windows does not provide the Linux Bash management menu or fail2ban integration.
For manual upgrades, stop the running process, back up the database and local
configuration, then replace program files while preserving your data.

OpenSSL is optional when you need to create or manage certificates yourself.
The repository no longer bundles a fixed third-party OpenSSL installer.
Windows distributions and their checksums are available from:
https://slproweb.com/products/Win32OpenSSL.html
Choose a supported Windows x64 distribution and follow its installation guide.

For a local development certificate, after installing OpenSSL:
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout localhost.key -out localhost.crt
Keep the generated private key outside version control.

Releases and version history:
https://github.com/yuchen0x1/Xray-server-management-panel/releases
https://github.com/yuchen0x1/Xray-server-management-panel/blob/main/docs/releases/release-notes.md
