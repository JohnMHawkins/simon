import usb.core
import usb.util
import time

#Simon LIght Controller

class DMXFixture:

    # string defining the type (e.g. "7-chan PAR" )
    fixtureType = ""

    # what is the DMX Address (1-512)
    dmxAddr = 1

    # how many channels does this fixture take
    numChan = 0

    # name : just a descriptive string, like "lefthand-spot"
    name = ""


    def __init__(self, numChan, fixType, name=""):
        self.numChan = numChan
        self.fixtureType = fixType
        self.name = name

    def setDMXAddr(self, dmxAddr):
        if dmxAddr >= 1 and dmxAddr <= 512:
            self.dmxAddr = dmxAddr

    # returns the highest address used by this fixture
    def lastDMXAddr(self):
        return self.dmxAddr + self.numChan

    def makeCommand(self,cmd,data):
        # this needs to be overridden by derrived classes
        return {'chanId': 0, 'data' : None}


class DMX_7_ChanPAR(DMXFixture):

    def __init__(self,name=""):
        DMXFixture.__init__(self, 7, "7-chan PAR")


    def makeCommand(self, cmd, data):
        ret = {'chanId':0, 'data':None}
        if cmd == "irgb":
            # sets intensity, red, green, blue, data is "i.r.g.b"
            dbytes = data[0].split(".")
            ret['chanId'] = 0
            ret['data'] = bytearray(int(dbytes[0]),int(dbytes[1]),int(dbytes[2]),int(dbytes[3]))
        elif cmd == "strobe":
            # sets strobe level.  data is 0-255 strobe intensity
            ret['chanId'] = 4
            ret['data'] = bytearray(int(dbytes))
        elif cmd == "off":
            # set mode to manual and then turn everything off
            ret['chanId'] = 0
            ret['data'] = bytearray(0,0,0,0,0,0,0)
        elif cmd == "mode":
            # sets the mode, data is modetype.modelevel.trigger]
            dbytes = data[0].split(".")
            if dbytes[0] == "select":
                mIdx = 10
            elif dbytes[0] == "shade":
                mIdx = 60
            elif dbytes[0] == "pulse":
                mIdx = 110
            elif dbytes[0] == "hard":
                mIdx = 160
            elif dbytes[0] == "sound":
                mIdx = 210
            else:
                # assume "manual"
                mIdx = 0
                dbytes[1] = "0"

            ch5val = mIdx = int(dbytes[1])
            ch6val = int(dbytes[2])    
            ret['chanId'] = 5
            ret['data'] = bytearray([ch5val, ch6val])
            

        return ret



class DMXController:

    # this is the device itself
    dev = None
    cfg = None
    bmReqType = usb.util.CTRL_TYPE_VENDOR | usb.util.CTRL_RECIPIENT_DEVICE | usb.util.CTRL_OUT

    # fixtures in this universe
    # { DMXFixture}
    fixtures = []


    def __init__(self):
        #find the device
        self.dev = usb.core.find(idVendor=0x16c0)
        print "Initializding DMXController"
        print self.dev
        print "---------------------------"
        self.dev.set_configuration()
        self.cfg = self.dev.get_active_configuration()
        

    def addFixture(self, fixture):
        self.fixtures.append(fixture)
        # TBD - go through self.fixtures and make sure there is no overlap?


    # returns the next open DMX address
    def getNextAvailAddr(self):
        nextId = 1
        for f in self.fixtures:
            if f.lastDMXAddr + 1 > nextId:
                nextId = f.lastDMXAddr + 1
        return nextid

    

    #
    # fixtureId is the index of the fixture (0,1,2 etc)
    # chanId is the channel on the fixture
    # data is the bytearray of data
    #
    def sendCommandX(self, fixtureId, chanId, data):
        numBytes = len(data)
        if fixtureId < len(self.fixtures):
            fix = self.fixtures[fixtureId]
            if (chanId + numBytes) <= fix['fixture'].numChan:
                wIdx = fix['fixture'].dmxAddr + chanId
                # send the command - we always use type = 2, multi channel, even if only sending one channel
                self.dev.ctrl_transfer(self.bmReqType, 2, wValue=numBytes, wIndex=wIdx, data_or_wLength=data)
            else:
                print "ERROR: trying to send too many channels for the device"
                print "Fixture: [" + str(fixtureId) + "] " + fix['name']
                print "fixture.numChan=" + str(fix['numChan'])
                print "chanId, bytes = " + str(chanId) + "," + str(numBytes)
        else:
            print "ERROR:  bad fixture id: " + str(fixtureId)


    # asks the DMXFixture to format a command and then sends it to the fixture
    # /{fixid}/{cmd}/{data}  e.g /1/irgb/255.255.255.0  tells a PAR to set full brightness yellow
    def sendCommand(self, cmdstring):
        cmdparts = cmdstring.split("/")
        if len(cmdparts) > 2:
            idx = int(cmdparts[1]) - 1
            if idx < len(self.fixtures):
                fix = self.fixtures[idx]
                cmd = fix.makeCommand(cmdparts[2], cmdparts[3:])
                if cmd.data != None:
                    wIdx = fix.dmxAddr + cmd.chanId
                    numBytes = len(cmd.data)
                    self.dev.ctrl_transfer(self.bmReqType, 2, wValue=numBytes, wIndex=wIdx, data_or_wLength=cmd.data)


