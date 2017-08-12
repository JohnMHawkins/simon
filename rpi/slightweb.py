import web
import json
import usb.core
import usb.util
import time
import subprocess


# stuff from slight


#find the device
dev = usb.core.find(idVendor=0x16c0)
print dev
dev.set_configuration()
cfg = dev.get_active_configuration()
intf = cfg[0,0]
bmReqType = usb.util.CTRL_TYPE_VENDOR | usb.util.CTRL_RECIPIENT_DEVICE | usb.util.CTRL_OUT


def sendCommand(fixtureId, startChan, cv):
    print dev.ctrl_transfer(bmReqType, 2, wValue=len(cv), wIndex=fixtureId + startChan, data_or_wLength=cv)


def turnOff(fixture):
    setMode(fixture, MODE_MANUAL, 0, 0)
    cv = bytearray([0,0,0,0])
    sendCommand(fixture,0,cv)
    
def setRGB(fixture,i, r, g, b):
    cv = bytearray([i,r,g,b])
    sendCommand(fixture, 0, cv)

def strobe(fixture, i):
    cv = bytearray([i])
    sendCommand(fixture, 4,cv)

MODE_MANUAL = 0
MODE_SELECT = 10
MODE_SHADE = 60
MODE_PULSE = 110
MODE_TRANS = 160
MODE_SOUND = 210

def setMode(fixture, mode, intensity, speed):
    ch5 = mode + intensity
    cv = bytearray([ch5, speed])
    sendCommand(fixture, 5, cv)
    

    

#turn light off until we're told otherwise

fixtures = [0,7]
settings = []

for f in fixtures:
    turnOff(f)
    setRGB(f, 255,255,255,255)

    #strobe off
    iStrobe = 0
    strobe(f, iStrobe)

    setting = { 'd': 255, 'r' : 0, 'g' : 0, 'b': 0 }
    settings.append(setting)
    


fx = 0

'''
bGo = True
while bGo:
    c = raw_input("command:")
    if c == 'q':
        for f in fixtures:
            turnOff(f)
        bGo = False
    elif c == 'r':
        s = settings[fx]
        s['r'] = 255
        s['g'] = 0
        s['b'] = 0
        setRGB(fixtures[fx], s['d'],s['r'],s['g'],s['b'])
    elif c == 'g':
        s = settings[fx]
        s['r'] = 0
        s['g'] = 255
        s['b'] = 0
        setRGB(fixtures[fx], s['d'],s['r'],s['g'],s['b'])
    elif c == 'b':
        s = settings[fx]
        s['r'] = 0
        s['g'] = 0
        s['b'] = 255
        setRGB(fixtures[fx], s['d'],s['r'],s['g'],s['b'])
    elif c == 'c':
        s = settings[fx]
        s['r'] = 0
        s['g'] = 255
        s['b'] = 255
        setRGB(fixtures[fx], s['d'],s['r'],s['g'],s['b'])
    elif c == 'p':
        s = settings[fx]
        s['r'] = 255
        s['g'] = 0
        s['b'] = 255
        setRGB(fixtures[fx], s['d'],s['r'],s['g'],s['b'])
    elif c == 'y':
        s = settings[fx]
        s['r'] = 255
        s['g'] = 255
        s['b'] = 0
        setRGB(fixtures[fx], s['d'],s['r'],s['g'],s['b'])
    elif c == 'w':
        s = settings[fx]
        s['r'] = 255
        s['g'] = 255
        s['b'] = 255
        setRGB(fixtures[fx], s['d'],s['r'],s['g'],s['b'])
    if c == "d":
        s = settings[fx]
        s['d'] = s['d'] - 75
        if s['d'] < 0:
            s['d'] = 255
        setRGB(fixtures[fx], s['d'],s['r'],s['g'],s['b'])
    elif c == 's':
        iStrobe = iStrobe + 75
        if iStrobe > 255 :
            iStrobe = 0
        strobe(fixtures[fx], iStrobe)
    elif c == "1":
        setMode(fixtures[0], MODE_MANUAL,0,0)
        setMode(fixtures[1], MODE_MANUAL,0,0)
    elif c == "2":
        setMode(fixtures[0], MODE_SELECT,25,127)
        setMode(fixtures[1], MODE_SELECT,25,127)
    elif c == "3":
        setMode(fixtures[0], MODE_SHADE,25,127)
        setMode(fixtures[1], MODE_SHADE,25,127)
    elif c == "4":
        setMode(fixtures[0], MODE_PULSE,25,127)
        setMode(fixtures[1], MODE_PULSE,25,127)
    elif c == "5":
        setMode(fixtures[0], MODE_TRANS,25,127)
        setMode(fixtures[1], MODE_TRANS,25,127)
    elif c == "6":
        setMode(fixtures[0], MODE_SOUND,25,127)
        setMode(fixtures[1], MODE_SOUND,25,127)

    elif c[0:1] == "f":
        fx = int(c[1:]) - 1
        print "set fixture to " + str(fixtures[fx])
        
    elif c == 'o':
        turnOff(fixtures[fx])

    time.sleep(0.5)
    
'''


# webserver stuff


urls = (
    '/web/stop', 'web_stop',
    '/web/shutdown', 'web_shutdown',
    '/dmx/assign', 'assign_fixture',
    '/dmx/cmd/(.*)', 'dmx_cmd'

)

app = web.application(urls, globals())


class assign_fixture:

    def GET(self):
        params = web.input()
        print "assigning"
        print params
        return "1"


class dmx_cmd:

    def GET(self, cmd):
        params = web.input()
        print "dmx_cmd:" + cmd
        print params
        if cmd == 'off':
            for f in fixtures:
                turnOff(f)
        elif cmd == 'rgb':
            #setRGB(fixtures[fx], s['d'],s['r'],s['g'],s['b'])
            cv = {'d' : 0, 'r' : 0, 'g' : 0, 'b': 0}
            if 'i' in params:
                cv['d'] = int(params['i'])
            if 'r' in params:
                cv['r'] = int(params['r'])
            if 'g' in params:
                cv['g'] = int(params['g'])
            if 'b' in params:
                cv['b'] = int(params['b'])
            for fx in fixtures:
                setRGB(fx, cv['d'],cv['r'],cv['g'],cv['b'])
            return cv



class web_stop:
    def GET(self):
        app.stop()
        return "shutting down"

class web_shutdown:
    def GET(self):
        command = '/usr/bin/sudo /sbin/shutdown now'
        process = subprocess.Popen(command.split(), stdout=subprocess.PIPE)
        output = process.communicate()[0]
        print output
        

if __name__ == "__main__":
    print "running the app"
    app.run()
