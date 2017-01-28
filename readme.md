#README#

This is a browser-based editor to allow contributors access to edit speicific files or directories on a server.

##Current State##
1. Uses editormd for editing md files and previewing there content.
1. Editing and saving files is working.
1. Has a crude implementation permissions and authentication that doesn't work for real world applications.
1. Uses go-bindata (https://github.com/jteeuwen/go-bindata) for binding resources to binary. Use the following commands to get bindata on a machine:
    go get github.com/jteeuwen/go-bindata/...
    go get github.com/elazarl/go-bindata-assetfs/...


##TODO:##
###Make choices on the follow technologies:###

Authentication and user storage methods

Javascript text editor

Language for templating (golang templates?)

