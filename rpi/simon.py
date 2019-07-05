import sys
# PI
import smbus
from gpiozero import LED
import pygame.mixer
from threading import Timer
import threading
import random
import time

'''
sudo apt-get install python-smbus python3-smbus python-dev python3-dev i2c-tools
sudo apt install python3-gpiozero
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

sounds = [[None, None, None],[None, None, None],[None, None, None],[None, None, None],[None, None, None],[None, None, None]]
audioch1 = None

# PI
i2cbus = smbus.SMBus(1)
ardAddr = 0x03


'''
4, 24, 17, 25, 27, 8, 22, 7, 10, 12, 9, 16, 11, 20, 5, 21



Pins for light relays:
7      - GPIO 4      - power on:  
8, 10  - GPIO 14, 15 - Red:  button, spot
11, 12 - GPIO 17, 18 - Green:  button, spot
13, 15 - GPIO 27, 22 - Blue:  button, spot
16, 18 - gpio 23, 24 - Yellow: button, spot

Center:  3x wrgby = 15
19, 21, 22 - GPIO 10, 9, 25  - w0, 1, 2 
23, 24, 26 - GPIO 11, 8, 7   - r0, 1, 2  
29, 31, 32 - GPIO 5, 6, 12   - g0, 1, 2
33, 35, 36 - GPIO 13, 19, 16 - b0, 1, 2
37, 38, 40 - GPIO 26, 20, 21 - y0, 1, 2

'''

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
        #           print("+++ turn on " + str(self.pin))
        pass

    def off(self):
        #print("--- turn off " + str(self.pin))
        pass
'''
def pollInput():
    global buttonPushed
    bGet = True
    while bGet:
        b = input("push a button: ")
        if b == "r":
            buttonPushed = SIMON_RED
        elif b == "g":
            buttonPushed = SIMON_GREEN
        elif b == "B":
            buttonPushed = SIMON_BLUE
        elif b == "y":
            buttonPushed = SIMON_YELLOW
        elif b == "c":
            buttonPushed = SIMON_CENTER
            
        LOG("input set buttonPushed to " + str(buttonPushed))
            

def useTestInput():
    t = threading.Thread(None, pollInput, name="testub", daemon=True)
    t.start()

##
##############################################

def setupLights():

    # light pins
    global PowerLight
    global ButtonLights
    global SpotLights
    global CenterLights
    
    PowerLight = LED(14)                 # GPIO 14
    
    ButtonLight[SIMON_RED] = LED(4)     # GPIO 4
    ButtonLight[SIMON_GREEN] = LED(17)  # GPIO 17
    ButtonLight[SIMON_BLUE] = LED(27)   # GPIO 27
    ButtonLight[SIMON_YELLOW] = LED(22) # GPIO 22

    SpotLights[SIMON_RED] = LED(15)     #
    SpotLights[SIMON_GREEN] = LED(18)   #
    SpotLights[SIMON_BLUE] = LED(23)    #
    SpotLights[SIMON_YELLOW] = LED(24)  #

    CenterLights[SIMON_WHITE][0] = LED(6)   # GPIO 6
    CenterLights[SIMON_WHITE][1] = LED(13)  # GPIO 13
    CenterLights[SIMON_WHITE][2] = LED(19)  # GPIO 19

    CenterLights[SIMON_RED][0] = LED(10)    # GPIO 10
    CenterLights[SIMON_RED][1] = LED(8) 
    CenterLights[SIMON_RED][2] = LED(7)

    CenterLights[SIMON_GREEN][0] = LED(9)   # GPIO 9
    CenterLights[SIMON_GREEN][1] = LED(26)   
    CenterLights[SIMON_GREEN][2] = LED(12)

    CenterLights[SIMON_BLUE][0] = LED(11)   # GPIO 11
    CenterLights[SIMON_BLUE][1] = LED(25)
    CenterLights[SIMON_BLUE][2] = LED(16)

    CenterLights[SIMON_YELLOW][0] = LED(5) # GPIO 5
    CenterLights[SIMON_YELLOW][1] = LED(20)
    CenterLights[SIMON_YELLOW][2] = LED(21)

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

# all lights except power
def allLightsOff():
    for i in range(SIMON_RED, SIMON_LAST + 1):
        #print("light off " + str(i))
        ButtonLight[i].off()
        SpotLights[i].off()
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
    CenterLights[color][0].off()
    CenterLights[color][1].off()
    CenterLights[color][2].off()

def colorOn(color):
    LOG("colorOn : " + str(color))
    if color != SIMON_WHITE:
        ButtonLight[color].on()
        SpotLights[color].on()
    CenterLights[color][0].on()
    CenterLights[color][1].on()
    CenterLights[color][2].on()

def centerWhiteOff(idx):
    #LOG("centerWhiteOff : " + str(idx))
    CenterLights[SIMON_CENTER][idx].off()

def centerWhiteOn(idx):
    #LOG("centerWhiteOn : " + str(idx))
    CenterLights[SIMON_CENTER][idx].on()


def flashColor(color, dur, bSound):
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
    t1 = Timer(syncDelay, colorOn, [color])
    t2 = Timer(syncDelay + dur, colorOff, [color])
    #colorOn(color)
    t1.start()
    t2.start()

##
############################################################3

def startPlayerTimer():
    LOG("startPlayerTimer")
    timerClick = 1.0
    t1 = Timer(timerClick, centerWhiteOff, [2])
    t2 = Timer(2*timerClick, centerWhiteOff, [1])
    t3 = Timer(3*timerClick, centerWhiteOff, [0])
    tDone = Timer(3*timerClick, gotoState, [STATE_PLAYER])
    for i in range(0,3):
        centerWhiteOn(i)
        
    t1.start()
    t2.start()
    t3.start()
    tDone.start()



def LOG(msg):
    print(msg)

def sendDMX(dest, cmd, data):
    # TODO
    pass


def setupArduinos():
    # TODO
    LOG("setupArduionos")
    setupLights()
    setupSounds()
    if bTestMode:
        pass
        #useTestInput()



def checkArduions():
    # ask arduionos if they are ready
    # TODO
    if False: #bTestMode:
        return True
    else:
        #TODO
        try:
            sendCommandToAll(CMD_ZERO, [], True)
            buttons = i2cbus.read_i2c_block_data(ardAddr, 1)
            LOG(buttons)
            return buttons != None and len(buttons) > 5
        except:
            return False

attractSequence = [SIMON_WHITE, SIMON_RED, SIMON_GREEN, SIMON_BLUE, SIMON_YELLOW]
#attractSequence = [SIMON_RED, SIMON_GREEN, SIMON_BLUE, SIMON_YELLOW]
attractDur = 1.0

def onAttractModeStep():
    global gameState
    global attractStep
    global attractSequence
    global attractDur
    LOG("onAttractModeStep")
    if gameState == STATE_ATTRACT:
        # show the next color
        if attractStep >= len(attractSequence):
            attractStep = 0

            # for testing without a center button
            #newGame()
            #bWaitForState = True

        attractThread = Timer(attractDur, onAttractModeStep, [])
        flashColor(attractSequence[attractStep], attractDur, False)
        attractStep = attractStep + 1
        attractThread.start()
    else:
      attractThread = None


def startAttractMode():
    global gameState
    global attractStep
    global attractThread
    global attractSequence
    global attractDur
    LOG("Start attract mode")
    attractStep = 0
    attractThread = Timer(attractDur, onAttractModeStep, [])
    flashColor(attractSequence[attractStep], attractDur, False)
    attractStep = attractStep + 1
    attractThread.start()

def newGame():
    global curSequence
    global curStep
    global bGameOver
    # Turn off all the lights, reset the sequence, then wait one second and start
    sendCommandToAll(CMD_ZERO, [], True)
    t = Timer(1.0, gotoState, [STATE_COMPUTER])
    allLightsOff()
    curSequence = []
    curStep = 0
    bGameOver = False
    t.start()
    


def addNewColor():
    global curSequence
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
    gotoState(STATE_SHOW)


def showNextColor(dur):
    global curSequence
    global curStep
    global bGameOver
    l = len(curSequence)
    if curStep == len(curSequence):
        if bGameOver:
            gotoState(STATE_ATTRACT)
        else:
            # reset curstep for player
            curStep = 0
            print("reset curstep to zero")
            t = Timer(1.0, gotoState, [STATE_START_TIMER])
            t.start()
            #gotoState(STATE_START_TIMER)
    else:
        t = Timer(dur, showNextColor, [dur])
        flashColor(curSequence[curStep], dur, True)
        curStep = curStep + 1
        t.start()


def showSequence():
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
    showNextColor(dur)

def showError():
    LOG("Showerror!!!  Game Over")
    t1 = Timer(0.6, flashColor, [SIMON_ERROR, 0.5, True])
    t2 = Timer(1.2, flashColor, [SIMON_ERROR, 0.5, True])
    flashColor(SIMON_ERROR, 0.5, True)        
    t1.start()
    t2.start()


def makePlayersChoice():
    # TODO
    global buttonPushed
    readButtons()
    if bTestMode:
        # set button pushed to last color
        if curStep < len(curSequence):
            if len(curSequence) > 3:
                buttonPushed = -1   # end the game
            else:
                buttonPushed = curSequence[curStep]
    LOG("Player pushed: " + str(buttonPushed))


# check to see if the player got it right
def evaluateChoice():
    global buttonPushed
    global curSequence
    global curStep
    LOG("+++ evaluateChoice for curstep " + str(curStep) + ", seq = " + str(curSequence))
    bOk = True
    if curStep >= len(curSequence):
        # huh, this is an error, shouldn't get here
        LOG("ERROR: evaluating choice beyond the curSequence")
        gotoState(STATE_ATTRACT)    
        return
    
    LOG("Evaluating:  pushed = " + str(buttonPushed) + ", expected = " + str(curSequence[curStep]))
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

        t = Timer(delay, gotoState, [nextState])
        flashColor(buttonPushed, 0.5, True)
        t.start()
    else:
        # failed!!!!
        t = Timer(2.5, gotoState, [STATE_GAMEOVER])
        showError()
        t.start()



        

        
def sendCommand(buttonId, cmd, data):
    # TODO

    '''
    # to send a command to the arduino:
    i2cbus.write_byte_data(ardAddr, 0, cmd)
    '''
    pass




def sendCommandToAll(cmd, data, includeCenter):
    if includeCenter == True:
        iStart = SIMON_CENTER
    else:
        iStart = SIMON_CENTER + 1

    # question:  do we need to set a timer delay for each command, or can we send them as fast as possible?
    for i in range (iStart, SIMON_LAST + 1):
        sendCommand(i, cmd, data)



def readButtons():
    global buttonPushed
    # TODO
    '''
    # read from i2c bus
    buttons = i2cbus.read_i2c_block_data(ardAddr, 1)
    buttons[SIMON_RED] = weight of red button, etc

    '''
    if bTestMode == False or gameState == STATE_ATTRACT:
        buttonPushed = -1
        # read from i2c bus
        try:
            buttons = i2cbus.read_i2c_block_data(ardAddr, 1)
            #LOG(buttons)
            maxWeight = 0
            for i in range(SIMON_CENTER, SIMON_LAST+1):
                if buttons[i] > maxWeight:
                    maxWeight = buttons[i]
                    buttonPushed = i
        except:
            LOG("error reading i2c")
            pass    
    #LOG("button pushed = " + str(buttonPushed))
    return buttonPushed



def gotoState(newState):
    global gameState
    global bWaitForState
    global attractThread
    LOG("gamestate chainging from " + str(gameState) + " to " + str(newState))
    if newState == STATE_ATTRACT:
        startAttractMode()
        sendCommandToAll(CMD_ATTRACT, [], True)
    elif attractThread != None:
        attractThread.cancel()
        attractThread = None


    gameState = newState
    bWaitForState = False




def loop():
    global gameState
    global bWaitForState
    global bGameOver
    global PowerLight
    #LOG("Loop: " + str(gameState))

    if bWaitForState:
        return

    if gameState == STATE_INIT:
        setupArduinos()
        sendDMX(SIMON_CENTER, DMX_OFF, None)
        gotoState(STATE_WAIT)

    elif gameState == STATE_WAIT:
        bArduinosReady = False
        while bArduinosReady == False:
            bArduinosReady = checkArduions()

        PowerLight.on()
        gotoState(STATE_ATTRACT)   

    elif gameState == STATE_ATTRACT:
        if readButtons() == SIMON_CENTER:
            # we can start
            LOG("button pushed :: start game")
            sendCommand(SIMON_CENTER, CMD_COMPUTER, {})
            newGame()
            bWaitForState = True
        #else:
        #    time.sleep(0.25)

    elif gameState == STATE_BEGIN:
        pass
    elif gameState == STATE_COMPUTER:
        addNewColor()
        #bWaitForState = True

    elif gameState == STATE_SHOW:
        LOG("State show")
        # be patient while simon is showing the player the sequence
        showSequence()
        bWaitForState = True

    elif gameState == STATE_START_TIMER:
        # Start the time
        startPlayerTimer()
        gotoState(STATE_TIMER)

    elif gameState == STATE_TIMER:
        # waiting for the time to run down
        readButtons()
        
    elif gameState == STATE_PLAYER:
        makePlayersChoice()
        gotoState(STATE_EVALUATE)

    elif gameState == STATE_EVALUATE:
        # checking to see if the player got it right
        evaluateChoice()
        bWaitForState = True
        pass

    elif gameState == STATE_GAMEOVER:
        # showing the game over sequence, then returning to Attract
        bGameOver = True
        gotoState(STATE_REPLAY)


    elif gameState == STATE_REPLAY:
        showSequence()
        bWaitForState = True

    elif gameState == STATE_TEST:
        pass
    
    elif gameState == STATE_CHECK_BUTTONS:
        pass


def main(argv):
    global bTestMode
    if len(sys.argv) > 1:
        if sys.argv[1] == "test":
            bTestMode = True
            print("Test mode - will simulate player hitting buttons")
        else:
            bTestMode = False

    bGo = True
    while bGo:
        loop()


if __name__ == "__main__":
    main(sys.argv)