#include <TimerThree.h>
#include <FastLED.h>
#include "HX711.h"

// vars for serial communication
String inputString = "";
String simonCommand = "";
boolean stringComplete = false;
String deviceName = "unknown";
String deviceInput = "unknown";



///////////////////////////////////////////////////////////
// identification

//Arduino Identification Pins (set to ground)
int PIN_CENTER = 51;

int PIN_BLUEGREEN = 52;
int PIN_REDYELLOW = 53;

int PIN_BLUE = 50;
int PIN_RED = 49;
int PIN_GREEN = 48;
int PIN_YELLOW = 47;

void setupIDPin() {
  pinMode(PIN_CENTER, INPUT);
  pinMode(PIN_BLUEGREEN, INPUT);
  pinMode(PIN_REDYELLOW, INPUT);
  pinMode(PIN_BLUE, INPUT);
  pinMode(PIN_GREEN, INPUT);
  pinMode(PIN_RED, INPUT);
  pinMode(PIN_YELLOW, INPUT);

  digitalWrite(PIN_CENTER, HIGH);
  digitalWrite(PIN_BLUEGREEN, HIGH);
  digitalWrite(PIN_REDYELLOW, HIGH);
  digitalWrite(PIN_BLUE, HIGH);
  digitalWrite(PIN_GREEN, HIGH);
  digitalWrite(PIN_RED, HIGH);
  digitalWrite(PIN_YELLOW, HIGH);
}

int getLoadCellPin(String input)
{
  if (deviceName == "REDYELLOW")
  {
    if (input == "RRR")
    {
      return 1;
    }
    else if (input == "YYY")
    {
      return 2;
    }
  }
  else if (deviceName == "BLUEGREEN")
  {
    if (input == "BBB")
    {
      return 1;
    }
    else if (input == "GGG")
    {
      return 2;
    }
  }
  return 1;
}


void simonAssignDeviceName() {

  //deviceName = getValue(inputString, ':', 1);

  int valCENTER = digitalRead(PIN_CENTER);
  int valBLUEGREEN = digitalRead(PIN_BLUEGREEN);
  int valREDYELLOW = digitalRead(PIN_REDYELLOW);
  int valBLUE = digitalRead(PIN_BLUE);
  int valGREEN = digitalRead(PIN_GREEN);
  int valRED = digitalRead(PIN_RED);
  int valYELLOW = digitalRead(PIN_YELLOW);

  if (valCENTER == LOW) {
    deviceName = "CENTER";
  } else if (valBLUEGREEN == LOW) {
    deviceName = "BLUEGREEN";
  } else if (valREDYELLOW == LOW) {
    deviceName = "REDYELLOW";
  } else if (valBLUE == LOW) {
    deviceName = "BLUE";
  } else if (valGREEN == LOW) {
    deviceName = "GREEN";
  } else if (valRED == LOW) {
    deviceName = "RED";
  } else if (valYELLOW == LOW) {
    deviceName = "YELLOW";
  } else {
    deviceName = "ERROR:UNKNOWN";
  }

  Serial.print("NAME:");
  Serial.println(deviceName);

}



void setup() {
  // put your setup code here, to run once:

  //setup arduino identification pins
  setupIDPin();

  //init serial
  Serial.begin (9600);
  Serial.setTimeout (50);
  inputString.reserve(200);  

  //init name
  simonAssignDeviceName();

}

void loop() {
  // put your main code here, to run repeatedly:





  noInterrupts();
  if (stringComplete) {
    parseCommand();
  }
  interrupts();

}



////////////////////////////////////////////////////////////////////////////////////
// 
// Serial port stuff
//

// send data out
void simonOutput (String outputString) {
  Serial.print(deviceInput);
  Serial.print(":");
  Serial.println(outputString);
}

// parsing vars
String namePart = "";
String commandPart = "";

// device names:
// CCC = CENTER
// RRR = RED
// GGG = GREEN
// BBB = BLUE
// YYY = YELLOW
//
// commands
// 111 = READ_BUTTONS
// 222 = FLASHCOLOR
// 333 = GS_TIMER
// 444 = GS_COMPUTER
// 555 = GS_ATTRACT

void parseCommand () {

  //our giant if statement to compare all commands

  //assign device nmae
  if (inputString.substring(0, 4) == "NAME" ) {
    //assign device nmae
    simonAssignDeviceName();

  } else {
    //simonOutput(inputString);
    //get deviceInputName and process command
    namePart = getValue(inputString, ':', 0);
    switch ( namePart[0]){
      case 'C':
        deviceInput = "CCC";
        break;
      case 'R':
        deviceInput = "RRR";
        break;
      case 'G':
        deviceInput = "GGG";
        break;
      case 'B':
        deviceInput = "BBB";
        break;
      case 'Y':
        deviceInput = "YYY";
        break;     
      default:
        deviceInput = "XXX";
        break;
    }
    commandPart = getValue(inputString, ':', 1);
    switch ( commandPart[0]) {
      case '1':
        simonCommand = "READ_BUTTONS";
        break;
      case '2':
        simonCommand = "FLASHCOLOR";
        break;
      case '3':
        simonCommand = "GS_TIMER";
        break;
      case '4':
        simonCommand = "GS_COMPUTER";
        break;
      case '5':
        simonCommand = "GS_ATTRACT";
        break;
      default:
        simonCommand = "NOCMD";
        break;
    }
    processSimonCommand();
  }

  // clear inputString, deviceInput & commandString
  inputString = "";
  deviceInput = "";
  simonCommand = "";
  stringComplete = false;

}

void processSimonCommand() {

  if (simonCommand == "FLASHCOLOR") {
    String colorRGB = getValue(inputString, ':', 2);

    String colorRed = getValue(colorRGB, '-', 0);
    String colorGreen = getValue(colorRGB, '-', 1);
    String colorBlue = getValue(colorRGB, '-', 2);

    String flashTime = getValue(inputString, ':', 3);

    String flashAmmount = getValue(inputString, ':', 4);
    simonOutput(deviceInput + ":" + colorRed + ":" + colorGreen + ":" + colorBlue + " FLASHTIME:" + flashTime + " FLASHAMMOUNT:" + flashAmmount);

    CRGB _color;
    _color.r = colorRed.toInt();
    _color.g = colorGreen.toInt();
    _color.b = colorBlue.toInt();

    FlashColor(_color, flashAmmount.toInt(), flashTime.toInt(),  flashTime.toInt(), 0);

  } else if (simonCommand == "GS_ATTRACT") {
    simonOutput (simonCommand);
    Rainbow(0);

  } else if (simonCommand == "GS_COMPUTER") {
    simonOutput (simonCommand);
    Rainbow(0);
    

  } else if (simonCommand == "GS_TIMER") {
    String colorRGB = getValue(inputString, ':', 2);

    String colorRed = getValue(colorRGB, '-', 0);
    String colorGreen = getValue(colorRGB, '-', 1);
    String colorBlue = getValue(colorRGB, '-', 2);

    String colorTime = getValue(inputString, ':', 3);
    simonOutput(deviceInput + ":" + colorRed + ":" + colorGreen + ":" + colorBlue + " TIME:" + colorTime);

    CRGB _color;
    _color.r = colorRed.toInt();
    _color.g = colorGreen.toInt();
    _color.b = colorBlue.toInt();

    CountDown(_color,  colorTime.toInt(), 0);


  } else if (simonCommand == "READ_BUTTONS") {
    //create test button data, replace with real button function
    int tmpDeviceNum = getLoadCellPin(deviceInput);
    if (tmpDeviceNum == 1)
    {
      simonOutput ("WWW:" + String(getWeight(1)));
    }
    else if (tmpDeviceNum == 2)
    {
      simonOutput ("WWW:" + String(getWeight(2)));
    }

  } else {
    simonOutput ("NOCMD:" + simonCommand);
  }

}

//
// serialEvent grabs all input from serial port and creates an inputString.
//
void serialEvent () {
  // called when serial port has data
  while (Serial.available() && !stringComplete) {
    char inChar = (char)Serial.read();

    // if newline, setflag so main loop can do something about it
    if (inChar == '\n') {
      stringComplete = true;
    } else {
      //add it to inputString
      inputString += inChar;
    }
  }

}


/*
   takes a string and separates it based on a given character and returns The item between the separating character
*/
String getValue(String data, char separator, int index)
{
  int found = 0;
  int strIndex[] = { 0, -1 };
  int maxIndex = data.length() - 1;

  for (int i = 0; i <= maxIndex && found <= index; i++) {
    if (data.charAt(i) == separator || i == maxIndex) {
      found++;
      strIndex[0] = strIndex[1] + 1;
      strIndex[1] = (i == maxIndex) ? i + 1 : i;
    }
  }
  return found > index ? data.substring(strIndex[0], strIndex[1]) : "";
}


/////////////////////////////////////////////////////////////////////////////////
//
// Load cell stuff
//

int getWeight(int button)
{
  return button * 1000;

}


/*
--------------------------------------------------------------------------------------------------------------------------
 * This function flashes a color on the LEDs
 * <color> is the color the light will flash
 * <annount> is the ammount of times it will flash
 * <timePerFlash> is the time the color is lit
 * <timeBetweenFkash> is the time the color is off (if your only flashing once it dose not matter what you set this to)
 * <ledStrip> is the stip it will flash on (0 or 1)
--------------------------------------------------------------------------------------------------------------------------
 */
void FlashColor(CRGB color, int ammount, int timePerFlash, int timeBetweenFlash, int ledStrip)
{

}

/*
--------------------------------------------------------------------------------------------------------------------------
 * This function counts down
 * <color> is the color the light
 * <_time> is the ammount of time it will take
 * <ledStrip> is the stip it will count down on
--------------------------------------------------------------------------------------------------------------------------
 */
void CountDown(CRGB color, int _time, int ledStrip)
{
  
}

/*
--------------------------------------------------------------------------------------------------------------------------
 * This function makes the leds Rainbow
 * <ledStrip> is the stip it will light on (0 or 1)
--------------------------------------------------------------------------------------------------------------------------
 */
void Rainbow(int ledStrip)
{

}

/*
--------------------------------------------------------------------------------------------------------------------------
 * This function changes the color of the LEDs to a solid color
 * <color> is the color the light will be
 * <ledStrip> is the stip it will color (0 or 1)
--------------------------------------------------------------------------------------------------------------------------
 */
void ChangeColor(CRGB color, int ledStrip)
{

}

