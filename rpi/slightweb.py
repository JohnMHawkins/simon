import web
import json
import slight


'''

api endpoints:
/dmx/assign?ch=<ch#>&type=fixturetype

    assigns <ch#> as the start channel for fixture type,e.g. 7ChanPAR

/dmx/raw/<ch#>/<0-255>

    sends 0-255 to the channel within the universe

/dmx/cmd/raw?fix=<fixid>&ch=<chNum>&val=<0-255>

    sends a raw number to a channel of a fixture.
    fixid = id of the fixture
    chNum = relative channel number (1-x) on fixture
    val = the value to send

/dmx/cmd/rgb?fix=<fixid>&r=<0-255>&b=<0-255>&g=<0-255>&i=<0-255>

    sets color for the fixture identified by fixid
    r = red value
    g = green value
    b = blue value
    i = intensity/dimmer value
    (each of these is optional)

/dmx/cmd/mode?fix=<fixid>&mode=<mode>&x=<x>&y=<y>

    sets the mode on the fixture identified by fixid
    mode = name of the mode, depends on fixture
    x = optional numeric value, depends on mode
    y = 2nd optional numeric value, depends on mode


/dmx/cmd/off?fixid=<fixid>

    turns the specified fixture off/blackout




'''

urls = (
    '/', 'index',
    '/dmx/assign', 'assign_fixture',
    '/cmd', 'cmd_all',
    #'/cmd/(.*)', 'cmd_fixture',
    '/cmd/(.*)/(.*)', 'cmd_fixture_foo'
)

devx = slight.DMXController()

fixtures = []


#adds the fixure, returns the id
def addFixture(fixtype, chan):

    newfix = {'type':fixtype, 'controller' : None, 'startchan' : chan, 'endchan' : -1 }

    #look up the controller, calc last channel

    fixnum = len(fixtures) + 1
    fixtures.append(newfix)
    return fixnum
    
'''
/dmx/assign?ch=<ch#>&type=fixturetype

    assigns <ch#> as the start channel for fixture type,e.g. 7ChanPAR
'''
class assign_fixture:

    def GET(self):
        params = web.input()


        return "1"
        


class index:
    def GET(self):
        return "Hello, world!"


class cmd_all:
    def GET(self):
        ip = web.input()
        print "command sent to all " + str(ip)
        return "Command sent to all "  + str(ip)

class cmd_fixture:
    def GET(self, fixture):
        print "command sent to fixture " + fixture
        return "command sent to fixture " + fixture


class cmd_fixture_foo:
    def GET(self, fixture, foo):
        print "command sent to fixture " + fixture + " of " + foo
        return "command sent to fixture " + fixture + " of " + foo

    
if __name__ == "__main__":
    app = web.application(urls, globals())
    app.run()
