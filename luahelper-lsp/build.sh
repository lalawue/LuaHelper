CGO_ENABLED=0 GOARCH=amd64 go build
mv luahelper-lsp ./../luahelper-vscode/server/maclualsp

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
mv luahelper-lsp ./../luahelper-vscode/server/linuxlualsp

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
mv luahelper-lsp.exe ./../luahelper-vscode/server/lualsp.exe
