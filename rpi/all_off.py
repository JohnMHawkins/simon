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

bTestMode = False

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

# commands
CMD_TEST = 1
CMD_ZERO = 2
CMD_ATTRACT = 3
CMD_COMPUTER = 4

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

DMXLightOn(SIMON_BLUE)
allLightsOff()

# our queue
cmdq = queue.PriorityQueue(0)
nextCmd = None