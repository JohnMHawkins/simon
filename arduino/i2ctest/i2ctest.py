import smbus

'''
sudo apt-get install python-smbus python3-smbus python-dev python3-dev i2c-tools

don't forget to enable the i2c protocol in pi config
'''

ardAdr = 0x03
CMD_TEST = 1
BTN_CENTER = 0
BTN_RED = 1
BTN_GREEN = 2
BTN_BLUE = 3
BTN_YELLOW = 4



def sendCmd(cmd):
    bus.write_byte_data(ardAdr, 0, cmd)


def readButton(btn):
    btns = bus.read_i2c_block_data(ardAdr, 1)
    #print(btns[btn])
    print(btns)


bus = smbus.SMBus(1)

bGo = True

while bGo:
    k = input("t=test, c=center, r=red, g=green, b=blue, y=yellow, q = quit - :")
    if k == "t":
        sendCmd(CMD_TEST)
    elif k == "c":
        readButton(BTN_CENTER)
    elif k == "r":
        readButton(BTN_RED)
    elif k == "g":
        readButton(BTN_GREEN)
    elif k == "b":
        readButton(BTN_BLUE)
    elif k == "y":
        readButton(BTN_YELLOW)
    elif k == "q":
        print("Got quit")
        bGo = False
