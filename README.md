# Bell scheduler(web-based scheduled-relay)
## A web-based bell scheduler(web-based scheduled-relay) on Raspberry Pi, coding in golang
This application was created for an existing warehouse bell system, upgrade from manual to automatic control

![diagram](https://raw.githubusercontent.com/qienhuang/bell_scheduler/master/snapshots/animation_bell_scheduler.gif)

### The application:

[bell_scheduler.go](https://github.com/qienhuang/bell_scheduler/blob/master/bell_scheduler.go)  # Written in Go and
uses third party packages:
- Gin framework as web server([github.com/gin-gonic/gin](https://github.com/gin-gonic/gin))
- go-rpio for relay control ([github.com/stianeikeland/go-rpio](https://github.com/stianeikeland/go-rpio))
- cron as scheduler ([github.com/robfig/cron](https://github.com/robfig/cron))

### The app runs on console:
<img src="https://raw.githubusercontent.com/qienhuang/bell_scheduler/master/snapshots/console.png" >


### The web interface:
<img src="https://raw.githubusercontent.com/qienhuang/bell_scheduler/master/snapshots/web_page.png" style="height=100% width=100%"> 


[index.html](https://github.com/qienhuang/bell_scheduler/blob/master/templates/index.html)  # html/jQuery
For users update the bell schedule on PC or mobile phone

### #Build the application
```
sudo go build
```
### #Install as a service
```
sudo cp systemctl/bell_scheduler.service /etc/systemd/system/
sudo systemctl enable bell_scheduler.service
sudo systemctl start bell_scheduler.service
```
### #Login to the web page

http://raspberrypi  (your local raspberry pi hostname)

### #A live demo web site:

[http://demo1.newddns.com:58000](http://demo1.newddns.com:58000)

User: admin, Password: password
