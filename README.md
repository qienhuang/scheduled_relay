# Bell scheduler
## A web-based bell scheduler on Raspberry Pi, coding in golang
This application was created for an existing warehouse bell system, upgrade from manual to automatic control

![diagram](https://raw.githubusercontent.com/qienhuang/bell_scheduler/master/snapshots/animation_bell_scheduler.gif)

### The application:

[bell_scheduler.go](https://github.com/qienhuang/bell_scheduler/blob/master/bell_scheduler.go)  # Written in Go
Using third party packages:
- Gin framework as web server([github.com/gin-gonic/gin](https://github.com/gin-gonic/gin))
- go-rpio for relay control ([github.com/stianeikeland/go-rpio](https://github.com/stianeikeland/go-rpio))
- cron as scheduler ([github.com/robfig/cron](https://github.com/robfig/cron))

The app runs on console:
<img src="https://raw.githubusercontent.com/qienhuang/bell_scheduler/master/snapshots/console.png" width="896" height="556">


The web interface:
<img src="https://raw.githubusercontent.com/qienhuang/bell_scheduler/master/snapshots/web_page.png" width="887" height="1054">


[index.html](https://github.com/qienhuang/bell_scheduler/blob/master/templates/index.html)  # html/jQuery
For users update the bell schedule on PC or mobile phone
