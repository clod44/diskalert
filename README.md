_go_ implementation of https://github.com/clod44/disk_monitor

currently does not have web push notification system

compile from the code:

```
go run build.go
```

transfer to vbox

```
scp ./dist/diskalert pacs@192.168.66.250:/home/pacs/diskalert/
```

put it in a folder for now. run it with `./diskalert`

go into the `ssl` folder and transfer the `diskalert.crt` file to the client device and install it

### client crt installation

this is required since we are working with a self signed crt and only way for push web notifications to work is to have a secured connection at the time of subscription.

- double click to the .crt file to open the windows crt installer.
- select "local machine"
- select the custom certification installation path and in the opened window, select "trusted root authorities" or whatever
- click done
