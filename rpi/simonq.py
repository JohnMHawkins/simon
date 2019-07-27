import sys

import pygame.mixer
from threading import Timer
import threading
import queue
import random
import time

from DMXEnttecPro import Controller
from DMXEnttecPro.utils import get_port_by_serial_number, get_port_by_product_id


'''
sudo apt-get install python-smbus python3-smbus python-dev python3-dev i2c-tools
sudo apt install python3-gpiozero
sudo pip3 install DMXEnttecPro
'''

bTestMode = True

SIMON_CENTER = 0
SIMON_NONE   = 0
SIMON_WHITE  = 0
SIMON_RED    = 1
SIMON_GREEN  = 2
SIMON_BLUE   = 3
SIMON_YELLOW = 4
SIMON_LAST   = 4
SIMON_ERROR  = 5

# Game states
STATE_INIT     = 0  # booting up and initializing
STATE_WAIT     = 1  # 
STATE_ATTRACT  = 2  # running attract sequence, waiting for START button
STATE_BEGIN    = 3  # new game received, counting down to start
STATE_COMPUTER = 4  # computer is thinking about new sequence
STATE_SHOW     = 5  # showing the new sequence
STATE_START_TIMER = 6
STATE_TIMER    = 7  # timer is running down, waiting for player input
STATE_PLAYER   = 8  # timer finished, figure out what button player pushed
STATE_EVALUATE = 9  # evaluating player input
STATE_GAMEOVER = 10  # player failed
STATE_REPLAY   = 11 # replaying last sequence
STATE_TEST     = 12 # testing all lights and audio
STATE_CHECK_BUTTONS = 13 # testing button input, will light whatever button the player is pushing



#dmx commands
DMX_OFF = 0
DMX_ON  = 1

gameState = STATE_INIT
bWaitForState = False
curSequence = []
curStep = 0
bGameOver = False
buttonPushed = -1
attractStep = 0
attractThread = None
retry = 0

sounds = [[None, None, None],[None, None, None],[None, None, None],[None, None, None],[None, None, None],[None, None, None]]
audioch1 = None


def LOG(msg):
    print(msg)


# PI
#import smbus
#from gpiozero import LED
#i2cbus = smbus.SMBus(1)
#ardAddr = 0x03


try:
    dmxUsb = get_port_by_product_id(24577)
    dmx = Controller(dmxUsb)
except:
    LOG("No DMX device detected")
    dmx = None


PowerLight = None
ButtonLight = [None, None, None, None, None, None] 
SpotLights = [None, None, None, None, None, None] 
CenterLights = [[None, None, None], [None, None, None], [None, None, None], [None, None, None], [None, None, None], [None, None, None]] 

############################################
#test code for not on Pi
# PI

class LED():

    def __init__(self, gpio):
        self.pin =  gpio

    def on(self):
        #           print("+++ turn on " + str(self.pin))
        pass

    def off(self):
        #print("--- turn off " + str(self.pin))
        pass


##
##############################################

def setupLights():

    # light pins
    global PowerLight
    global ButtonLights
    global SpotLights
    global CenterLights
    
    PowerLight = LED(11)                 
    
    ButtonLight[SIMON_RED] = LED(26)     
    ButtonLight[SIMON_GREEN] = LED(19)  
    ButtonLight[SIMON_BLUE] = LED(13)   
    ButtonLight[SIMON_YELLOW] = LED(6) 

    SpotLights[SIMON_RED] = LED(27)     
    SpotLights[SIMON_GREEN] = LED(22)   
    SpotLights[SIMON_BLUE] = LED(4)    
    SpotLights[SIMON_YELLOW] = LED(17)  

    CenterLights[SIMON_WHITE][0] = LED(10)   
    CenterLights[SIMON_WHITE][1] = LED(9)  
    CenterLights[SIMON_WHITE][2] = LED(5)  

    CenterLights[SIMON_RED][0] = LED(12)    
    CenterLights[SIMON_RED][1] = LED(14) 
    CenterLights[SIMON_RED][2] = LED(15)

    CenterLights[SIMON_GREEN][0] = LED(16)   
    CenterLights[SIMON_GREEN][1] = LED(18)   
    CenterLights[SIMON_GREEN][2] = LED(23)

    CenterLights[SIMON_BLUE][0] = LED(20)   
    CenterLights[SIMON_BLUE][1] = LED(24) 
    CenterLights[SIMON_BLUE][2] = LED(25) 

    CenterLights[SIMON_YELLOW][0] = LED(21) 
    CenterLights[SIMON_YELLOW][1] = LED(8)
    CenterLights[SIMON_YELLOW][2] = LED(7)

    # error is just RED
    ButtonLight[SIMON_ERROR] = ButtonLight[SIMON_RED]
    SpotLights[SIMON_ERROR] = SpotLights[SIMON_RED]   #
    CenterLights[SIMON_ERROR][0] = CenterLights[SIMON_RED][0]
    CenterLights[SIMON_ERROR][1] = CenterLights[SIMON_RED][1] 
    CenterLights[SIMON_ERROR][2] = CenterLights[SIMON_RED][2]


def setupSounds():
    #return
    global sounds
    global audioch1

    pygame.mixer.init()
    audioch1 = pygame.mixer.Channel(1)

    sounds[SIMON_RED][0] = pygame.mixer.Sound("audio/red_500.wav")
    sounds[SIMON_RED][1] = pygame.mixer.Sound("audio/red_250.wav")
    sounds[SIMON_RED][2] = pygame.mixer.Sound("audio/red_125.wav")

    sounds[SIMON_GREEN][0] = pygame.mixer.Sound("audio/green_500.wav")
    sounds[SIMON_GREEN][1] = pygame.mixer.Sound("audio/green_250.wav")
    sounds[SIMON_GREEN][2] = pygame.mixer.Sound("audio/green_125.wav")

    sounds[SIMON_BLUE][0] = pygame.mixer.Sound("audio/blue_500.wav")
    sounds[SIMON_BLUE][1] = pygame.mixer.Sound("audio/blue_250.wav")
    sounds[SIMON_BLUE][2] = pygame.mixer.Sound("audio/blue_125.wav")

    sounds[SIMON_YELLOW][0] = pygame.mixer.Sound("audio/yellow_500.wav")
    sounds[SIMON_YELLOW][1] = pygame.mixer.Sound("audio/yellow_250.wav")
    sounds[SIMON_YELLOW][2] = pygame.mixer.Sound("audio/yellow_125.wav")

    sounds[SIMON_ERROR][0] = pygame.mixer.Sound("audio/error_500.wav")


def playSound(color, dur):
    global sounds
    global audioch1
    LOG("playsound " + str(color))
    if color > 0 :
        audioch1.play(sounds[color][dur])

######################################################################3
## Light control methods


# DMX methods

# the DMX_xxx values below are the offsets we will use for each fixture.  
# DMX_ALL = any fixture that should show all colors
# DMX_RED = any fixture to show only Red
# DMX_GREEN = any fixture to show only Green
# DMX_BLUE = any fixture to show only Blue.
# DMX_YELLOW = any fixture to show only Yellow.
#
# the aXXX values are the settings on the fixture for seven channel mode, aXXX is for three channel mode
#
# the DMX_xxx value below will be used to set the channel as:
#   intensity = DMX_xxx + 0
#   red = DMX_xxx + 1
#   green = DMX_xxx + 2
#   blue = DMX_xxx + 3

# 3-channel mode (e.g. all = d001, red = d004, green = d007, blue = d010, yellow = d013 )
# in this mode, channel 1 is read, 
dmx7Chan = False
DMX_ALL = 0     # set fixture to d001
DMX_RED = 3     # set fixture to d004
DMX_GREEN = 6   # set fixture to d007
DMX_BLUE = 9    # set fixture to d010
DMX_YELLOW = 12 # set fixture to d013

# 7 - channel Mode (e.g. all= a001, red = a008, green = a015, blue = a022, yellow = a029)
# in this mode, channel 0 is intensity, so we set the aXXX value and the channel below to be the same for each fixture, as intensity is Start + 0
# and red is Start + 1 (so for red, we want to use a009)
dmx7Chan = True
dmxIntensity = 255
DMX_ALL = 1         # set fixture to a001
DMX_RED = 8         # set fixture to a008
DMX_GREEN = 15      # set fixture to a015
DMX_BLUE = 22       # set fixture to a022
DMX_YELLOW = 29     # set fixture to a029


def DMXLightOn(color):
    if dmx == None:
        return


    if color == SIMON_RED:
        if dmx7Chan:
            dmx.set_channel(DMX_ALL,dmxIntensity)
            dmx.set_channel(DMX_RED,dmxIntensity)
        dmx.set_channel(DMX_ALL + 1, 255)
        dmx.set_channel(DMX_ALL + 2, 0)
        dmx.set_channel(DMX_ALL + 3, 0)

        dmx.set_channel(DMX_RED + 1, 255)
        dmx.set_channel(DMX_RED + 2, 0)
        dmx.set_channel(DMX_RED + 3, 0)
    
    elif color == SIMON_GREEN:
        if dmx7Chan:
            dmx.set_channel(DMX_ALL,dmxIntensity)
            dmx.set_channel(DMX_GREEN,dmxIntensity)
        dmx.set_channel(DMX_ALL + 1, 0)
        dmx.set_channel(DMX_ALL + 2, 255)
        dmx.set_channel(DMX_ALL + 3, 0)

        dmx.set_channel(DMX_GREEN + 1, 0)
        dmx.set_channel(DMX_GREEN + 2, 255)
        dmx.set_channel(DMX_GREEN + 3, 0)

    elif color == SIMON_BLUE:
        if dmx7Chan:
            dmx.set_channel(DMX_ALL,dmxIntensity)
            dmx.set_channel(DMX_BLUE,dmxIntensity)
        dmx.set_channel(DMX_ALL + 1, 0)
        dmx.set_channel(DMX_ALL + 2, 0)
        dmx.set_channel(DMX_ALL + 3, 255)

        dmx.set_channel(DMX_BLUE + 1, 0)
        dmx.set_channel(DMX_BLUE + 2, 0)
        dmx.set_channel(DMX_BLUE + 3, 255)

    elif color == SIMON_YELLOW:
        if dmx7Chan:
            dmx.set_channel(DMX_ALL,dmxIntensity)
            dmx.set_channel(DMX_YELLOW,dmxIntensity)
        dmx.set_channel(DMX_ALL + 1, 255)
        dmx.set_channel(DMX_ALL + 2, 128)
        dmx.set_channel(DMX_ALL + 3, 0)

        dmx.set_channel(DMX_YELLOW + 1, 255)
        dmx.set_channel(DMX_YELLOW + 2, 128)
        dmx.set_channel(DMX_YELLOW + 3, 0)

    elif color == SIMON_ERROR:
        if dmx7Chan:
            dmx.set_channel(DMX_RED,dmxIntensity)
            dmx.set_channel(DMX_GREEN,dmxIntensity)
            dmx.set_channel(DMX_BLUE,dmxIntensity)
            dmx.set_channel(DMX_YELLOW,dmxIntensity)
        dmx.set_channel(DMX_ALL + 1, 255)
        dmx.set_channel(DMX_ALL + 2, 0)
        dmx.set_channel(DMX_ALL + 3, 0)

        dmx.set_channel(DMX_RED + 1, 255)
        dmx.set_channel(DMX_RED + 2, 0)
        dmx.set_channel(DMX_RED + 3, 0)

        dmx.set_channel(DMX_GREEN + 1, 255)
        dmx.set_channel(DMX_GREEN + 2, 0)
        dmx.set_channel(DMX_GREEN + 3, 0)

        dmx.set_channel(DMX_BLUE + 1, 255)
        dmx.set_channel(DMX_BLUE + 2, 0)
        dmx.set_channel(DMX_BLUE + 3, 0)

        dmx.set_channel(DMX_YELLOW + 1, 255)
        dmx.set_channel(DMX_YELLOW + 2, 0)
        dmx.set_channel(DMX_YELLOW + 3, 0)

    dmx.submit()


def DMXLightOff(color):
    if dmx == None:
        return

    if dmx7Chan:
        dmx.set_channel(DMX_ALL,0)
    dmx.set_channel(DMX_ALL + 1, 0)
    dmx.set_channel(DMX_ALL + 2, 0)
    dmx.set_channel(DMX_ALL + 3, 0)

    if color == SIMON_RED:
        if dmx7Chan:
            dmx.set_channel(DMX_RED,0)
        dmx.set_channel(DMX_RED + 1, 0)
        dmx.set_channel(DMX_RED + 2, 0)
        dmx.set_channel(DMX_RED + 3, 0)

    elif color == SIMON_GREEN:
        if dmx7Chan:
            dmx.set_channel(DMX_GREEN,0)
        dmx.set_channel(DMX_GREEN + 1, 0)
        dmx.set_channel(DMX_GREEN + 2, 0)
        dmx.set_channel(DMX_GREEN + 3, 0)

    elif color == SIMON_BLUE:
        if dmx7Chan:
            dmx.set_channel(DMX_BLUE,0)
        dmx.set_channel(DMX_BLUE + 1, 0)
        dmx.set_channel(DMX_BLUE + 2, 0)
        dmx.set_channel(DMX_BLUE + 3, 0)

    elif color == SIMON_YELLOW:
        if dmx7Chan:
            dmx.set_channel(DMX_YELLOW,0)
        dmx.set_channel(DMX_YELLOW + 1, 0)
        dmx.set_channel(DMX_YELLOW + 2, 0)
        dmx.set_channel(DMX_YELLOW + 3, 0)

    elif color == SIMON_ERROR:
        if dmx7Chan:
            dmx.set_channel(DMX_RED,0)
            dmx.set_channel(DMX_GREEN,0)
            dmx.set_channel(DMX_BLUE,0)
            dmx.set_channel(DMX_YELLOW,0)
        dmx.set_channel(DMX_RED + 1, 0)
        dmx.set_channel(DMX_RED + 2, 0)
        dmx.set_channel(DMX_RED + 3, 0)

        dmx.set_channel(DMX_GREEN + 1, 0)
        dmx.set_channel(DMX_GREEN + 2, 0)
        dmx.set_channel(DMX_GREEN + 3, 0)

        dmx.set_channel(DMX_BLUE + 1, 0)
        dmx.set_channel(DMX_BLUE + 2, 0)
        dmx.set_channel(DMX_BLUE + 3, 0)

        dmx.set_channel(DMX_YELLOW + 1, 0)
        dmx.set_channel(DMX_YELLOW + 2, 0)
        dmx.set_channel(DMX_YELLOW + 3, 0)

    dmx.submit()

# all lights except power
def allLightsOff():
    for i in range(SIMON_RED, SIMON_LAST + 1):
        #print("light off " + str(i))
        ButtonLight[i].off()
        SpotLights[i].off()
        DMXLightOff(i)
        for j in range(0, 3):
            CenterLights[i][j].off()
    for j in range(0, 3):
        #print("center off " + str(j))
        CenterLights[SIMON_WHITE][j].off()

    print("done all lights off")

def setPowerLight(bOn):
    if bOn:
        PowerLight.on()
    else:
        PowerLight.off()


def colorOff(color):
    LOG("colorOff : " + str(color))
    if color != SIMON_WHITE:
        ButtonLight[color].off()
        SpotLights[color].off()
        DMXLightOff(color)
    CenterLights[color][0].off()
    CenterLights[color][1].off()
    CenterLights[color][2].off()

def colorOn(color):
    LOG("colorOn : " + str(color))
    if color != SIMON_WHITE:
        ButtonLight[color].on()
        SpotLights[color].on()
        DMXLightOn(color)
    CenterLights[color][0].on()
    CenterLights[color][1].on()
    CenterLights[color][2].on()

def centerWhiteOff(idx):
    #LOG("centerWhiteOff : " + str(idx))
    CenterLights[SIMON_CENTER][idx].off()

def centerWhiteOn(idx):
    #LOG("centerWhiteOn : " + str(idx))
    CenterLights[SIMON_CENTER][idx].on()



print("hello")
setupLights()
setupSounds()


##################################################################3
# Arduino code
#

def checkArduions():
    # ask arduionos if they are ready
    if bTestMode:
        LOG("checkArduinos, test = true")
        return True
    else:
        try:
            LOG("about to try arduino")
            sendCommandToAll(CMD_ZERO, [], True)
            buttons = i2cbus.read_i2c_block_data(ardAddr, 1)
            LOG(buttons)
            return buttons != None and len(buttons) > 5
        except:
            return False


def ReadButtons():
    pass




def newGame(ts):
    global curSequence
    global curStep
    global bGameOver
    # Turn off all the lights, reset the sequence, then wait one second and start
    #sendCommandToAll(CMD_ZERO, [], True)
    AddCmd(ts + 1.0, CMD_GOTO_STATE, STATE_COMPUTER)
    allLightsOff()
    curSequence = []
    curStep = 0
    bGameOver = False
    


def addNewColor(ts):
    global curSequence
    global bWaitForState
    newcolor = random.randint(SIMON_RED, SIMON_YELLOW)
    # Don't allow more than three in a row
    l = len(curSequence)
    if l > 2:
        if curSequence[l-1] == newcolor and curSequence[l-2] == newcolor and curSequence[l-3] == newcolor:
            LOG("shifting so we don't have >3 in a row")
            adj = random.randint(1,3)
            newcolor = newcolor + adj
            if newcolor > SIMON_YELLOW:
                newcolor = newcolor - 4

    curSequence.append(newcolor)
    LOG(curSequence)
    AddCmd(ts + 0.1, CMD_GOTO_STATE, STATE_SHOW)
    bWaitForState = True



def showNextColor(ts, dur):
    global curSequence
    global curStep
    global bGameOver
    global bWaitForState
    l = len(curSequence)
    if curStep == len(curSequence):
        if bGameOver:
            AddCmd(ts + 1.0, CMD_GOTO_STATE, STATE_ATTRACT)
            bWaitForState = True
        else:
            # reset curstep for player
            curStep = 0
            print("reset curstep to zero")
            AddCmd(ts + 1.0, CMD_GOTO_STATE, STATE_START_TIMER)
    else:
        AddCmd(ts + dur, CMD_SHOW_NEXT_COLOR, dur)
        flashColor(ts, curSequence[curStep], dur, True)
        curStep = curStep + 1


def showSequence(ts):
    global curSequence
    global curStep
    global bGameOver
    LOG("+++ ShowSequence " + str(curSequence))
    l = len(curSequence)
    if bGameOver:
        dur = 0.5
    elif l < 7:
        dur = 0.5
    elif l < 13:
        dur = 0.25
    else:
        dur = 0.125

    curStep = 0
    showNextColor(ts, dur)


#################################################################
# our queue and queue functions, definitions, etc
#

cmdq = queue.PriorityQueue(0)
nextCmd = None

# adds a command to the queue sorted by ts, keeping nextCmd as the next one in time sequence
def AddCmd(cmdAt, cmd, data):
    global nextCmd
    global cmdq
    LOG("added command " + str(cmd))
    if nextCmd == None:
        nextCmd = {'ts': cmdAt, 'cmd' : cmd, 'data' : data}
    elif cmdAt < nextCmd['ts']:
        cmdq.put((nextCmd['ts'], nextCmd['cmd'], nextCmd['data']))
        nextCmd = {'ts': cmdAt, 'cmd' : cmd, 'data' : data}
    else:
        cmdq.put((cmdAt, cmd, data))

# gets the next command if it's ts is <= the specified ts
def GetCmd(ts):
    global nextCmd
    global cmdq
    retCmd = None
    if nextCmd != None and ts>= nextCmd['ts']:
        retCmd = nextCmd
        if cmdq.empty() == False:
            cmdAt, cmd, data = cmdq.get(False)
            nextCmd = {'ts': cmdAt, 'cmd' : cmd, 'data' : data}
            cmdq.task_done()
        else:
            nextCmd = None
    return retCmd
        
    
CMD_GOTO_STATE = 1
CMD_LIGHT_ON = 2
CMD_LIGHT_OFF = 3
CMD_PLAY_SOUND = 4
CMD_ATTRACT_STEP = 5
CMD_SHOW_NEXT_COLOR = 6

def DoGotoState(ts, newState):
    global gameState
    global bWaitForState
    global attractThread
    LOG("gamestate chainging from " + str(gameState) + " to " + str(newState))
    gameState = newState
    if newState == STATE_ATTRACT:
        startAttractMode(ts)

    bWaitForState = False

def DoLightOn(ts, data):
    colorOn(data)
    pass

def DoLightOff(ts, data):
    colorOff(data)
    pass

def DoPlaySound(ts, data):
    pass

def DoShowNextColor(ts, data):
    showNextColor(ts, data)

def HandleCommand(cmd):
    LOG("HandleCommand " + str(cmd['cmd']))
    if cmd['cmd'] == CMD_GOTO_STATE:
        DoGotoState(cmd['ts'], cmd['data'])
    elif cmd['cmd'] == CMD_LIGHT_ON:
        DoLightOn(cmd['ts'], cmd['data']) 
    elif cmd['cmd'] == CMD_LIGHT_OFF:
        DoLightOff(cmd['ts'], cmd['data']) 
    elif cmd['cmd'] == CMD_PLAY_SOUND:
        DoPlaySound(cmd['ts'], cmd['data']) 
    elif cmd['cmd'] == CMD_ATTRACT_STEP:
        DoAttractModeStep(cmd['ts']) 
    elif cmd['cmd'] == CMD_SHOW_NEXT_COLOR:
        DoShowNextColor(cmd['ts'], cmd['data'])
       


# returns True if it handled a command       
def PeekCmd(ts):
    cmd = GetCmd(ts)
    if cmd != None:
        HandleCommand(cmd)
        return True
    else:
        return False



def flashColor(ts, color, dur, bSound):
    LOG("flashColor " + str(color) + ", dur=" + str(dur))
    idx = 0
    if bSound == True:
        if dur < 0.2:
            idx = 2
        elif dur < 0.5:
            idx = 1
        playSound(color, idx)

    # we use a delay to syncronize sound and light
    syncDelay = 0.55
    AddCmd(ts + syncDelay, CMD_LIGHT_ON, color)
    AddCmd(ts + syncDelay + dur, CMD_LIGHT_OFF, color)


attractSequence = [SIMON_WHITE, SIMON_RED, SIMON_BLUE, SIMON_YELLOW, SIMON_GREEN]
attractDur = 1.0

def DoAttractModeStep(ts):
    global gameState
    global attractStep
    global attractSequence
    global attractDur
    LOG("onAttractModeStep " + str(gameState))
    if gameState == STATE_ATTRACT:
        # show the next color
        if attractStep >= len(attractSequence):
            attractStep = 0

            # for testing without a center button
            if bTestMode:
                newGame(ts)
                bWaitForState = True

        AddCmd(ts + attractDur, CMD_ATTRACT_STEP, {})
        flashColor(ts, attractSequence[attractStep], attractDur, False)
        attractStep = attractStep + 1


def startAttractMode(ts):
    global gameState
    global attractStep
    global attractThread
    global attractSequence
    global attractDur
    LOG("Start attract mode")
    attractStep = 0
    DoAttractModeStep(ts)




def loop():
    global gameState
    global bWaitForState
    global bGameOver
    global PowerLight
    global retry
    #LOG("Loop: " + str(gameState))
    ts = time.time()

    bPeek = True
    while bPeek:
        bPeek = PeekCmd(ts)

    if bWaitForState:
        return

    if gameState == STATE_INIT:     # 0
        #setupArduinos()
        AddCmd(ts + 0.1, CMD_GOTO_STATE, STATE_WAIT)
        bWaitForState = True
        #gotoState(STATE_WAIT)
        pass

    elif gameState == STATE_WAIT:   # 1
        bArduinosReady = False
        while bArduinosReady == False:
            bArduinosReady = checkArduions()

        PowerLight.on()
        AddCmd(time.time(), CMD_GOTO_STATE, STATE_ATTRACT)
        pass

    elif gameState == STATE_ATTRACT:    # 2
        #readButtons()
        #if buttonPushed == SIMON_CENTER:
        #    # we can start
        #    LOG("button pushed :: start game")
        #    sendCommand(SIMON_CENTER, CMD_COMPUTER, {})
        #    newGame()
        #    bWaitForState = True
        pass

    elif gameState == STATE_BEGIN: #3
        pass

    elif gameState == STATE_COMPUTER:  #4
        addNewColor(ts)
        
        
    elif gameState == STATE_SHOW:   #5
        LOG("State show")
        # be patient while simon is showing the player the sequence
        showSequence(ts)
        bWaitForState = True
        
    elif gameState == STATE_START_TIMER:
        # Start the time
        #startPlayerTimer()
        #gotoState(STATE_TIMER)
        pass

    elif gameState == STATE_TIMER:
        # waiting for the time to run down
        #readButtons()
        #retry = 0
        pass

    elif gameState == STATE_PLAYER:
        #if makePlayersChoice() or retry > 5:
        #    gotoState(STATE_EVALUATE)
        #else:
        #    retry = retry + 1
        pass

    elif gameState == STATE_EVALUATE:
        # checking to see if the player got it right
        #evaluateChoice()
        #bWaitForState = True
        pass

    elif gameState == STATE_GAMEOVER:
        # showing the game over sequence, then returning to Attract
        #bGameOver = True
        #gotoState(STATE_REPLAY)
        pass

    elif gameState == STATE_REPLAY:
        #showSequence()
        #bWaitForState = True
        pass

    elif gameState == STATE_TEST:
        pass
    
    elif gameState == STATE_CHECK_BUTTONS:
        pass


def main(argv):
    global bTestMode

    try:
        bGo = True
        while bGo:
            loop()
    except KeyboardInterrupt:
        allLightsOff()
        bGo = False

if __name__ == "__main__":
    main(sys.argv)