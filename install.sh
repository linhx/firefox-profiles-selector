

tee -a ~/.local/share/applications/firefox-profiles-selector.desktop <<EOF
[Desktop Entry]
Version=1.0
Name=Firefox profiles selector
Comment=Firefox profiles selector
Exec=bash -i -c '$(pwd)/firefox-profiles-selector %u'
Terminal=false
Type=Application
MimeType=text/html;text/xml;application/xhtml+xml;application/xml;application/vnd.mozilla.xul+xml;application/rss+xml;application/rdf+xml;image/gif;image/jpeg;image/png;x-scheme-handler/http;x-scheme-handler/https;
Icon=$(pwd)/icon.png
EOF

xdg-settings set default-web-browser firefox-profiles-selector.desktop
