<<<<<<< HEAD
import sys

import pygame.mixer
from threading import Timer
import threading
import queue
import random
import time

from DMXEnttecPro import Controller
from DMXEnttecPro.utils import get_port_by_serial_number, get_port_by_product_id

# Simon_sim.py :  simulates simon on a PC without arduino connections

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

import smbus
from gpiozero import LED
i2cbus = smbus.SMBus(1)
ardAddr = 0x03


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
'''
class LED():

    def __init__(self, gpio):
        self.pin =  gpio

    def on(self):
        print("+++ turn on " + str(self.pin))
        pass

    def off(self):
        print("--- turn off " + str(self.pin))
        pass

'''
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

    pygame.mixer.init(channels=5)
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
        try:
            audioch1.play(sounds[color][dur])
        except:
            LOG("playsound exception")

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

'''
# 3-channel mode (e.g. all = d001, red = d004, green = d007, blue = d010, yellow = d013 )
# in this mode, channel 1 is read, 
dmx7Chan = False
DMX_ALL = 0     # set fixture to d001
DMX_RED = 3     # set fixture to d004
DMX_GREEN = 6   # set fixture to d007
DMX_BLUE = 9    # set fixture to d010
DMX_YELLOW = 12 # set fixture to d013
'''

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

    try:
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
    except:
        LOG("Exception in dmx DMXLightOn")

def DMXLightOff(color):
    if dmx == None:
        return

    try:
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
    except:
        LOG("Exception in dmx DMXLightOff")


# all lights except power
def allLightsOff():
    for i in range(SIMON_RED, SIMON_LAST + 1):
        #print("light off " + str(i))
        try:
            ButtonLight[i].off()
            SpotLights[i].off()
            DMXLightOff(i)
            for j in range(0, 3):
                CenterLights[i][j].off()
        except:
            LOG("exception allLightsOff")
    for j in range(0, 3):
        try:
            #print("center off " + str(j))
            CenterLights[SIMON_WHITE][j].off()
        except:
            LOG("exception CentireLightw WHITE off")
    print("done all lights off")

def setPowerLight(bOn):
    if bOn:
        return
        PowerLight.on()
    else:
        try:
            PowerLight.off()
        except:
            LOG("Exception PowerLight.off")

def colorOn(color):
    #LOG("colorOn : " + str(color))
    try:
        if color == SIMON_ERROR:
            color = SIMON_RED
        if color != SIMON_WHITE:
            ButtonLight[color].on()
            SpotLights[color].on()
            DMXLightOn(color)
        CenterLights[color][0].on()
        CenterLights[color][1].on()
        CenterLights[color][2].on()
    except:
        LOG("exception colorOn : " + str(color))

def colorOff(color):
    try:
        if color == SIMON_ERROR:
            color = SIMON_RED
        #LOG("colorOff : " + str(color))
        if color != SIMON_WHITE:
            ButtonLight[color].off()
            SpotLights[color].off()
            DMXLightOff(color)
        CenterLights[color][0].off()
        CenterLights[color][1].off()
        CenterLights[color][2].off()
    except:
        LOG("Exeption colorOff : " + str(color))

def centerWhiteOn(idx):
    #LOG("centerWhiteOn : " + str(idx))
    try:
        CenterLights[SIMON_CENTER][idx].on()
    except:
        LOG("Exception EnterWhite on")

def centerWhiteOff(idx):
    #LOG("centerWhiteOff : " + str(idx))
    try:
        CenterLights[SIMON_CENTER][idx].off()
    except:
        LOG("Exception CenterWhite Off")


print("hello")
setupLights()
setupSounds()


##################################################################3
# Arduino code
#

def checkArduinios():
    # ask arduionos if they are ready
    if bTestMode:
        LOG("checkArduinos, test = true")
        return True
    else:
        try:
            #LOG("about to try arduino")
            #sendCommandToAll(CMD_ZERO, [], True)   
            buttons = i2cbus.read_i2c_block_data(ardAddr, 1)
            #LOG(buttons)
            return buttons != None and len(buttons) > 4
        except:
            LOG("Error checking arduinos")
            return False

buttons = []
signals = [0,0,0,0,0]
minweight = 5
minsignals = 18
numreadfails = 0
expectedColor = 0
numExpectedSignals = 0

def sendCommand(buttonId, cmd, data):
    # to send a command to the arduino:
    try:
        i2cbus.write_byte_data(ardAddr, 0, cmd)
    except:
        LOG("Exception writing to i2cbus : " + str(ardAddr) + " : cmd=" + str(cmd))


def clearbuttons():
    global buttons
    global signals
    global expectedColor
    global numExpectedSignals
    expectedColor = 0
    numExpectedSignals = 0
    for i in range(SIMON_CENTER, SIMON_LAST+1):
        signals[i] = 0  

def ReadButtons():
    global buttons
    global signals
    global numreadfails
    if bTestMode == False: 
        try:
            buttons = i2cbus.read_i2c_block_data(ardAddr, 1)
            for i in range (SIMON_CENTER+1, SIMON_LAST+1):
                if buttons[i] > minweight:
                    signals[i] = signals[i] + 1
                else:
                    signals[i] = 0
            #LOG(buttons[:5])
        except:
            numreadfails = numreadfails + 1
            LOG("=======================================")
            LOG("   ")
            LOG("exception reading buttons " + str(numreadfails))
            LOG("   ")
            LOG("=======================================")
    pass


def evalButtons():
    global buttonPushed
    global buttons
    global signals
    global gameState
    global expectedColor
    global numExpectedSignals
    if buttons != None and len(buttons) > SIMON_LAST:
        #LOG(buttons[:5])
        if gameState == STATE_ATTRACT:
            if buttons[SIMON_CENTER] > minweight:
                buttonPushed = SIMON_CENTER
                LOG(buttons[:5])
        else:
            maxWeight = minweight      # filter out low weights
            for i in range(SIMON_CENTER, SIMON_LAST+1):
                if buttons[i] > maxWeight:
                    if i == expectedColor:
                        buttonPushed = i
                        numExpectedSignals = numExpectedSignals + 1
                    if signals[i] > minsignals:
                        maxWeight = buttons[i]
                        buttonPushed = i
            

def newGame(ts):
    global curSequence
    global curStep
    global bGameOver
    global gameState
    # Turn off all the lights, reset the sequence, then wait one second and start
    #sendCommandToAll(CMD_ZERO, [], True)
    #AddCmd(ts + 1.0, CMD_GOTO_STATE, STATE_COMPUTER)
    LOG("Starting new game")
    ClearCmdQ()
    allLightsOff()
    curSequence = []
    curStep = 0
    bGameOver = False
    gameState = STATE_BEGIN
    startCountdownMode(ts)
    


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

    '''
    newcolor = SIMON_BLUE
    if l > 0:
        if curSequence[l-1] == SIMON_BLUE:
            newcolor = SIMON_GREEN
        elif curSequence[l-1] == SIMON_GREEN:
            newcolor = SIMON_RED
        elif curSequence[l-1] == SIMON_RED:
            newcolor = SIMON_YELLOW
    '''

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
    #LOG("showNextColor step " + str(curStep))
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
        AddCmd(ts + dur + 0.05, CMD_SHOW_NEXT_COLOR, dur)
        flashColor(ts, curSequence[curStep], dur, True)
        curStep = curStep + 1


def showSequence(ts):
    global curSequence
    global curStep
    global bGameOver
    LOG("+++ ShowSequence len=" + str(len(curSequence)) + " " + str(curSequence))
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



def startPlayerTimer(ts):
    global curSequence
    global expectedColor
    global numExpectedSignals
    LOG("startPlayerTimer")
    timerClick = 1.0
    
    if len(curSequence) > 12:
        timerClick = 0.5
    elif len(curSequence) > 6:
        timerClick = 0.75
    

    AddCmd(ts + timerClick, CMD_CENTERWHITE_OFF, 2)
    AddCmd(ts + 2*timerClick, CMD_CENTERWHITE_OFF, 1)
    AddCmd(ts + 3*timerClick, CMD_CENTERWHITE_OFF, 0)
    AddCmd(ts + 3*timerClick + 0.25, CMD_GOTO_STATE, STATE_PLAYER)
    for i in range(0,3):
        centerWhiteOn(i)
    buttonPushed = -1

    expectedColor = curSequence[curStep]
    numExpectedSignals = 0
        


def makePlayersChoice():
    global numreadfails
    global buttonPushed

    ReadButtons()
    if bTestMode:
        # set button pushed to last color
        if curStep < len(curSequence):
            if len(curSequence) > 15:
                buttonPushed = -1   # end the game
            else:
                LOG("Sequence Length " + str(len(curSequence)))
                buttonPushed = curSequence[curStep]
    else:
        evalButtons()
    LOG("Player pushed: " + str(buttonPushed) + ", readfails = " + str(numreadfails))
    numreadfails = 0


# check to see if the player got it right
def evaluateChoice(ts):
    global buttonPushed
    global curSequence
    global curStep
    global retry
    LOG("+++ evaluateChoice for curstep " + str(curStep) + ", seq = " + str(curSequence))
    bOk = True
    if curStep >= len(curSequence):
        # huh, this is an error, shouldn't get here
        LOG("ERROR: evaluating choice beyond the curSequence")
        gotoState(STATE_ATTRACT)    
        return
    
    LOG("Evaluating:  pushed = " + str(buttonPushed) + ", expected = " + str(curSequence[curStep])+ ", retries = " + str(retry))
    bOk = buttonPushed == curSequence[curStep]
    curStep = curStep + 1
    if bOk:
        nextState = STATE_COMPUTER
        if curStep < len(curSequence):
            # player is right so far, show the color and go to the next one
            nextState = STATE_START_TIMER
            delay = 1.5
        else:
            # player got it right, add another color
            nextState = STATE_COMPUTER
            delay = 2.5

        AddCmd(ts+delay, CMD_GOTO_STATE, nextState)
        flashColor(ts, buttonPushed, 0.5, True)
    else:
        # failed!!!!
        LOG("FAILED ------")
        LOG(signals)
        LOG("numExpected = " + str(numExpectedSignals))
        LOG("------------")
        AddCmd(ts+2.5, CMD_GOTO_STATE, STATE_GAMEOVER)
        showError(ts)

    # reset button selection    
    buttonPushed = -1



def showError(ts):
    LOG("Showerror!!!  Game Over")
    LOG("total numreadfails = " + str(numreadfails))
    AddCmd(ts + 0.2, CMD_FLASH_COLOR, [SIMON_ERROR, 0.5, True])
    AddCmd(ts + 0.8, CMD_FLASH_COLOR, [SIMON_ERROR, 0.5, True])
    AddCmd(ts + 1.4, CMD_FLASH_COLOR, [SIMON_ERROR, 0.5, True])




#################################################################
# our queue and queue functions, definitions, etc
#

cmdq = queue.PriorityQueue(0)
nextCmd = None

# adds a command to the queue sorted by ts, keeping nextCmd as the next one in time sequence
def AddCmd(cmdAt, cmd, data):
    global nextCmd
    global cmdq
    #LOG("added command " + str(cmd) + " " + str(data) + " at " + str(cmdAt))
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
        
def ClearCmdQ():
    global nextCmd
    global cmdq
    nextCmd = None
    while cmdq.empty() == False:
        cmdAt, cmd, data = cmdq.get(False)
        cmdq.task_done()


    
CMD_GOTO_STATE = 1
CMD_LIGHT_ON = 2
CMD_LIGHT_OFF = 3
CMD_PLAY_SOUND = 4
CMD_ATTRACT_STEP = 5
CMD_SHOW_NEXT_COLOR = 6
CMD_CENTERWHITE_ON = 7
CMD_CENTERWHITE_OFF = 8
CMD_FLASH_COLOR = 9
CMD_COUNTDOWN_STEP = 10

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
    if data == SIMON_ERROR:
        colorOn(SIMON_RED)
        colorOn(SIMON_GREEN)
        colorOn(SIMON_BLUE)
        colorOn(SIMON_YELLOW)
    else:
        colorOn(data)
    
def DoLightOff(ts, data):
    if data == SIMON_ERROR:
        colorOff(SIMON_RED)
        colorOff(SIMON_GREEN)
        colorOff(SIMON_BLUE)
        colorOff(SIMON_YELLOW)
    else:
        colorOff(data)
    
def DoPlaySound(ts, data):
    pass

def DoShowNextColor(ts, data):
    showNextColor(ts, data)

def HandleCommand(cmd):
    #LOG("HandleCommand " + str(cmd['cmd']))
    try:
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
        elif cmd['cmd'] == CMD_CENTERWHITE_ON:
            centerWhiteOn(cmd['data'])
        elif cmd['cmd'] == CMD_CENTERWHITE_OFF:
            centerWhiteOff(cmd['data'])
        elif cmd['cmd'] == CMD_FLASH_COLOR:
            flashColor(cmd['ts'], cmd['data'][0],cmd['data'][1],cmd['data'][2]  )
        elif cmd['cmd'] == CMD_COUNTDOWN_STEP:
            DoCoundownModeStep(cmd['ts'])
    except:
        LOG("HandleCommand exception for cmd " + str(cmd['cmd']) + " : data=" + str(cmd['data']))


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
            random.shuffle(attractSequence)


            # for testing without a center button
            if bTestMode:
                newGame(ts)
                bWaitForState = True
                return

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


countdownSequence = [SIMON_WHITE, SIMON_RED, SIMON_BLUE, SIMON_YELLOW, SIMON_GREEN]
counddownStep = 4

def DoCoundownModeStep(ts):
    global counddownStep
    global countdownSequence
    LOG("onCountdownModeStep " + str(gameState))
    if counddownStep >= len(countdownSequence):
        for i in range(0, len(countdownSequence)):
           colorOn(countdownSequence[i])
        counddownStep = len(countdownSequence) -1
    else:
        # turn off the next color
        colorOff(countdownSequence[counddownStep])
        counddownStep = counddownStep - 1

    if counddownStep > -1:
        AddCmd(ts + 0.6, CMD_COUNTDOWN_STEP, {})
    else:
        AddCmd(ts + 0.6, CMD_GOTO_STATE, STATE_COMPUTER)
        bWaitForState = True

def startCountdownMode(ts):
    global counddownStep
    global countdownSequence
    LOG("Start countdown mode")
    counddownStep = len(countdownSequence)
    AddCmd(ts + 0.6, CMD_COUNTDOWN_STEP, {})



lastRead = 0
MINREADDELTA = 0.05

def loop():
    global gameState
    global bWaitForState
    global bGameOver
    global PowerLight
    global retry
    global lastRead
    #LOG("Loop: " + str(gameState))
    ts = time.time()
    bRead = False
    if ts - lastRead > MINREADDELTA:
        bRead = True
        lastRead = ts

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
            bArduinosReady = checkArduinios()

        PowerLight.on()
        AddCmd(time.time(), CMD_GOTO_STATE, STATE_ATTRACT)
        pass

    elif gameState == STATE_ATTRACT:    # 2
        if bRead:
            ReadButtons()
        evalButtons()
        if buttonPushed == SIMON_CENTER:
            # we can start
            LOG("button pushed :: start game")
            #sendCommand(SIMON_CENTER, 2, {})
            newGame(ts)
            bWaitForState = True
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
        
    elif gameState == STATE_START_TIMER:  #6
        # Start the time
        startPlayerTimer(ts)
        AddCmd(time.time(), CMD_GOTO_STATE, STATE_TIMER)
        pass

    elif gameState == STATE_TIMER:   #7
        # waiting for the time to run down
        if bRead:
            ReadButtons()
        #retry = 0
        pass

    elif gameState == STATE_PLAYER:  #8
        ReadButtons()
        makePlayersChoice()
        AddCmd(time.time(), CMD_GOTO_STATE, STATE_EVALUATE)
        bWaitForState = True


    elif gameState == STATE_EVALUATE:  #9
        # checking to see if the player got it right
        evaluateChoice(ts)
        bWaitForState = True
        pass

    elif gameState == STATE_GAMEOVER:  #10
        # showing the game over sequence, then returning to Attract
        bGameOver = True
        AddCmd(time.time(), CMD_GOTO_STATE, STATE_REPLAY)
        bWaitForState = True
        pass

    elif gameState == STATE_REPLAY:  #11
        showSequence(ts)
        bWaitForState = True
        pass

    elif gameState == STATE_TEST:  #12 
        pass
    
    elif gameState == STATE_CHECK_BUTTONS:  #13
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
=======
import sys

import pygame.mixer
from threading import Timer
import threading
import queue
import random
import time

from DMXEnttecPro import Controller
from DMXEnttecPro.utils import get_port_by_serial_number, get_port_by_product_id

# Simon_sim.py :  simulates simon on a PC without arduino connections

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


import smbus
from gpiozero import LED
i2cbus = smbus.SMBus(1)
ardAddr = 0x03


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
'''
class LED():

    def __init__(self, gpio):
        self.pin =  gpio

    def on(self):
        print("+++ turn on " + str(self.pin))
        pass

    def off(self):
        print("--- turn off " + str(self.pin))
        pass

'''
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

    pygame.mixer.init(channels=5)
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

'''
# 3-channel mode (e.g. all = d001, red = d004, green = d007, blue = d010, yellow = d013 )
# in this mode, channel 1 is read, 
dmx7Chan = False
DMX_ALL = 0     # set fixture to d001
DMX_RED = 3     # set fixture to d004
DMX_GREEN = 6   # set fixture to d007
DMX_BLUE = 9    # set fixture to d010
DMX_YELLOW = 12 # set fixture to d013
'''

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
        return
        PowerLight.on()
    else:
        PowerLight.off()


def colorOn(color):
    #LOG("colorOn : " + str(color))
    if color == SIMON_ERROR:
        color = SIMON_RED
    if color != SIMON_WHITE:
        ButtonLight[color].on()
        SpotLights[color].on()
        DMXLightOn(color)
    CenterLights[color][0].on()
    CenterLights[color][1].on()
    CenterLights[color][2].on()

def colorOff(color):
    if color == SIMON_ERROR:
        color = SIMON_RED
    #LOG("colorOff : " + str(color))
    if color != SIMON_WHITE:
        ButtonLight[color].off()
        SpotLights[color].off()
        DMXLightOff(color)
    CenterLights[color][0].off()
    CenterLights[color][1].off()
    CenterLights[color][2].off()

def centerWhiteOn(idx):
    #LOG("centerWhiteOn : " + str(idx))
    CenterLights[SIMON_CENTER][idx].on()

def centerWhiteOff(idx):
    #LOG("centerWhiteOff : " + str(idx))
    CenterLights[SIMON_CENTER][idx].off()


print("hello")
setupLights()
setupSounds()


##################################################################3
# Arduino code
#

def checkArduinios():
    # ask arduionos if they are ready
    if bTestMode:
        LOG("checkArduinos, test = true")
        return True
    else:
        try:
            #LOG("about to try arduino")
            #sendCommandToAll(CMD_ZERO, [], True)   
            buttons = i2cbus.read_i2c_block_data(ardAddr, 1)
            #LOG(buttons)
            return buttons != None and len(buttons) > 4
        except:
            LOG("Error checking arduinos")
            return False

buttons = []
signals = [0,0,0,0,0]
minweight = 5
minsignals = 18
numreadfails = 0
expectedColor = 0
numExpectedSignals = 0

def sendCommand(buttonId, cmd, data):
    # to send a command to the arduino:
    i2cbus.write_byte_data(ardAddr, 0, cmd)


def clearbuttons():
    global buttons
    global signals
    global expectedColor
    global numExpectedSignals
    expectedColor = 0
    numExpectedSignals = 0
    for i in range(SIMON_CENTER, SIMON_LAST+1):
        signals[i] = 0  

def ReadButtons():
    global buttons
    global signals
    global numreadfails
    if bTestMode == False: 
        try:
            buttons = i2cbus.read_i2c_block_data(ardAddr, 1)
            for i in range (SIMON_CENTER+1, SIMON_LAST+1):
                if buttons[i] > minweight:
                    signals[i] = signals[i] + 1
                else:
                    signals[i] = 0
            #LOG(buttons[:5])
        except:
            numreadfails = numreadfails + 1
            LOG("=======================================")
            LOG("   ")
            LOG("exception reading buttons " + str(numreadfails))
            LOG("   ")
            LOG("=======================================")
    pass


def evalButtons():
    global buttonPushed
    global buttons
    global signals
    global gameState
    global expectedColor
    global numExpectedSignals
    if buttons != None and len(buttons) > SIMON_LAST:
        #LOG(buttons[:5])
        if gameState == STATE_ATTRACT:
            if buttons[SIMON_CENTER] > minweight:
                buttonPushed = SIMON_CENTER
                LOG(buttons[:5])
        else:
            maxWeight = minweight      # filter out low weights
            for i in range(SIMON_CENTER, SIMON_LAST+1):
                if buttons[i] > maxWeight:
                    if i == expectedColor:
                        buttonPushed = i
                        numExpectedSignals = numExpectedSignals + 1
                    if signals[i] > minsignals:
                        maxWeight = buttons[i]
                        buttonPushed = i
            

def newGame(ts):
    global curSequence
    global curStep
    global bGameOver
    global gameState
    # Turn off all the lights, reset the sequence, then wait one second and start
    #sendCommandToAll(CMD_ZERO, [], True)
    #AddCmd(ts + 1.0, CMD_GOTO_STATE, STATE_COMPUTER)
    LOG("Starting new game")
    ClearCmdQ()
    allLightsOff()
    curSequence = []
    curStep = 0
    bGameOver = False
    gameState = STATE_BEGIN
    startCountdownMode(ts)
    


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

    '''
    newcolor = SIMON_BLUE
    if l > 0:
        if curSequence[l-1] == SIMON_BLUE:
            newcolor = SIMON_GREEN
        elif curSequence[l-1] == SIMON_GREEN:
            newcolor = SIMON_RED
        elif curSequence[l-1] == SIMON_RED:
            newcolor = SIMON_YELLOW
    '''

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
    #LOG("showNextColor step " + str(curStep))
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
        AddCmd(ts + dur + 0.05, CMD_SHOW_NEXT_COLOR, dur)
        flashColor(ts, curSequence[curStep], dur, True)
        curStep = curStep + 1


def showSequence(ts):
    global curSequence
    global curStep
    global bGameOver
    LOG("+++ ShowSequence len=" + str(len(curSequence)) + " " + str(curSequence))
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



def startPlayerTimer(ts):
    global curSequence
    global expectedColor
    global numExpectedSignals
    LOG("startPlayerTimer")
    timerClick = 1.0
    
    if len(curSequence) > 12:
        timerClick = 0.5
    elif len(curSequence) > 6:
        timerClick = 0.75
    

    AddCmd(ts + timerClick, CMD_CENTERWHITE_OFF, 2)
    AddCmd(ts + 2*timerClick, CMD_CENTERWHITE_OFF, 1)
    AddCmd(ts + 3*timerClick, CMD_CENTERWHITE_OFF, 0)
    AddCmd(ts + 3*timerClick + 0.25, CMD_GOTO_STATE, STATE_PLAYER)
    for i in range(0,3):
        centerWhiteOn(i)
    buttonPushed = -1

    expectedColor = curSequence[curStep]
    numExpectedSignals = 0
        


def makePlayersChoice():
    global numreadfails
    global buttonPushed

    ReadButtons()
    if bTestMode:
        # set button pushed to last color
        if curStep < len(curSequence):
            if len(curSequence) > 15:
                buttonPushed = -1   # end the game
            else:
                LOG("Sequence Length " + str(len(curSequence)))
                buttonPushed = curSequence[curStep]
    else:
        evalButtons()
    LOG("Player pushed: " + str(buttonPushed) + ", readfails = " + str(numreadfails))
    numreadfails = 0


# check to see if the player got it right
def evaluateChoice(ts):
    global buttonPushed
    global curSequence
    global curStep
    global retry
    LOG("+++ evaluateChoice for curstep " + str(curStep) + ", seq = " + str(curSequence))
    bOk = True
    if curStep >= len(curSequence):
        # huh, this is an error, shouldn't get here
        LOG("ERROR: evaluating choice beyond the curSequence")
        gotoState(STATE_ATTRACT)    
        return
    
    LOG("Evaluating:  pushed = " + str(buttonPushed) + ", expected = " + str(curSequence[curStep])+ ", retries = " + str(retry))
    bOk = buttonPushed == curSequence[curStep]
    curStep = curStep + 1
    if bOk:
        nextState = STATE_COMPUTER
        if curStep < len(curSequence):
            # player is right so far, show the color and go to the next one
            nextState = STATE_START_TIMER
            delay = 1.5
        else:
            # player got it right, add another color
            nextState = STATE_COMPUTER
            delay = 2.5

        AddCmd(ts+delay, CMD_GOTO_STATE, nextState)
        flashColor(ts, buttonPushed, 0.5, True)
    else:
        # failed!!!!
        LOG("FAILED ------")
        LOG(signals)
        LOG("numExpected = " + str(numExpectedSignals))
        LOG("------------")
        AddCmd(ts+2.5, CMD_GOTO_STATE, STATE_GAMEOVER)
        showError(ts)

    # reset button selection    
    buttonPushed = -1



def showError(ts):
    LOG("Showerror!!!  Game Over")
    LOG("total numreadfails = " + str(numreadfails))
    AddCmd(ts + 0.2, CMD_FLASH_COLOR, [SIMON_ERROR, 0.5, True])
    AddCmd(ts + 0.8, CMD_FLASH_COLOR, [SIMON_ERROR, 0.5, True])
    AddCmd(ts + 1.4, CMD_FLASH_COLOR, [SIMON_ERROR, 0.5, True])




#################################################################
# our queue and queue functions, definitions, etc
#

cmdq = queue.PriorityQueue(0)
nextCmd = None

# adds a command to the queue sorted by ts, keeping nextCmd as the next one in time sequence
def AddCmd(cmdAt, cmd, data):
    global nextCmd
    global cmdq
    #LOG("added command " + str(cmd) + " " + str(data) + " at " + str(cmdAt))
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
        
def ClearCmdQ():
    global nextCmd
    global cmdq
    nextCmd = None
    while cmdq.empty() == False:
        cmdAt, cmd, data = cmdq.get(False)
        cmdq.task_done()


    
CMD_GOTO_STATE = 1
CMD_LIGHT_ON = 2
CMD_LIGHT_OFF = 3
CMD_PLAY_SOUND = 4
CMD_ATTRACT_STEP = 5
CMD_SHOW_NEXT_COLOR = 6
CMD_CENTERWHITE_ON = 7
CMD_CENTERWHITE_OFF = 8
CMD_FLASH_COLOR = 9
CMD_COUNTDOWN_STEP = 10

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
    if data == SIMON_ERROR:
        colorOn(SIMON_RED)
        colorOn(SIMON_GREEN)
        colorOn(SIMON_BLUE)
        colorOn(SIMON_YELLOW)
    else:
        colorOn(data)
    
def DoLightOff(ts, data):
    if data == SIMON_ERROR:
        colorOff(SIMON_RED)
        colorOff(SIMON_GREEN)
        colorOff(SIMON_BLUE)
        colorOff(SIMON_YELLOW)
    else:
        colorOff(data)
    
def DoPlaySound(ts, data):
    pass

def DoShowNextColor(ts, data):
    showNextColor(ts, data)

def HandleCommand(cmd):
    #LOG("HandleCommand " + str(cmd['cmd']))
    try:
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
        elif cmd['cmd'] == CMD_CENTERWHITE_ON:
            centerWhiteOn(cmd['data'])
        elif cmd['cmd'] == CMD_CENTERWHITE_OFF:
            centerWhiteOff(cmd['data'])
        elif cmd['cmd'] == CMD_FLASH_COLOR:
            flashColor(cmd['ts'], cmd['data'][0],cmd['data'][1],cmd['data'][2]  )
        elif cmd['cmd'] == CMD_COUNTDOWN_STEP:
            DoCoundownModeStep(cmd['ts'])
    except:
        LOG("HandleCommand exception for cmd " + str(cmd['cmd']) + " : data=" + str(cmd['data']))


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
            random.shuffle(attractSequence)


            # for testing without a center button
            if bTestMode:
                newGame(ts)
                bWaitForState = True
                return

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


countdownSequence = [SIMON_WHITE, SIMON_RED, SIMON_BLUE, SIMON_YELLOW, SIMON_GREEN]
counddownStep = 4

def DoCoundownModeStep(ts):
    global counddownStep
    global countdownSequence
    LOG("onCountdownModeStep " + str(gameState))
    if counddownStep >= len(countdownSequence):
        for i in range(0, len(countdownSequence)):
           colorOn(countdownSequence[i])
        counddownStep = len(countdownSequence) -1
    else:
        # turn off the next color
        colorOff(countdownSequence[counddownStep])
        counddownStep = counddownStep - 1

    if counddownStep > -1:
        AddCmd(ts + 0.6, CMD_COUNTDOWN_STEP, {})
    else:
        AddCmd(ts + 0.6, CMD_GOTO_STATE, STATE_COMPUTER)
        bWaitForState = True

def startCountdownMode(ts):
    global counddownStep
    global countdownSequence
    LOG("Start countdown mode")
    counddownStep = len(countdownSequence)
    AddCmd(ts + 0.6, CMD_COUNTDOWN_STEP, {})



lastRead = 0
MINREADDELTA = 0.05

def loop():
    global gameState
    global bWaitForState
    global bGameOver
    global PowerLight
    global retry
    global lastRead
    #LOG("Loop: " + str(gameState))
    ts = time.time()
    bRead = False
    if ts - lastRead > MINREADDELTA:
        bRead = True
        lastRead = ts

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
            bArduinosReady = checkArduinios()

        PowerLight.on()
        AddCmd(time.time(), CMD_GOTO_STATE, STATE_ATTRACT)
        pass

    elif gameState == STATE_ATTRACT:    # 2
        if bRead:
            ReadButtons()
        evalButtons()
        if buttonPushed == SIMON_CENTER:
            # we can start
            LOG("button pushed :: start game")
            #sendCommand(SIMON_CENTER, 2, {})
            newGame(ts)
            bWaitForState = True
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
        
    elif gameState == STATE_START_TIMER:  #6
        # Start the time
        startPlayerTimer(ts)
        AddCmd(time.time(), CMD_GOTO_STATE, STATE_TIMER)
        pass

    elif gameState == STATE_TIMER:   #7
        # waiting for the time to run down
        if bRead:
            ReadButtons()
        #retry = 0
        pass

    elif gameState == STATE_PLAYER:  #8
        ReadButtons()
        makePlayersChoice()
        AddCmd(time.time(), CMD_GOTO_STATE, STATE_EVALUATE)
        bWaitForState = True


    elif gameState == STATE_EVALUATE:  #9
        # checking to see if the player got it right
        evaluateChoice(ts)
        bWaitForState = True
        pass

    elif gameState == STATE_GAMEOVER:  #10
        # showing the game over sequence, then returning to Attract
        bGameOver = True
        AddCmd(time.time(), CMD_GOTO_STATE, STATE_REPLAY)
        bWaitForState = True
        pass

    elif gameState == STATE_REPLAY:  #11
        showSequence(ts)
        bWaitForState = True
        pass

    elif gameState == STATE_TEST:  #12 
        pass
    
    elif gameState == STATE_CHECK_BUTTONS:  #13
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
>>>>>>> cde12e7335eec054f4c5ac296eed0742f65a6f1f
    main(sys.argv)