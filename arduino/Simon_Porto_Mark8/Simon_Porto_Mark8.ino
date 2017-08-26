#include <Arduino.h>

#include <TimerThree.h>
#include <FastLED.h>
#include "HX711.h"

//enums
enum CallBack {FLASHCOLOR, COUNTDOWN, UNDEF};

enum State {RAINBOW, SOLID, COUNTING};

//init leds
//light settings 1
#define DATA_PIN1    2
#define CLK_PIN1     3
#define LED_TYPE1    APA102
#define COLOR_ORDER1 BGR
#define NUM_LEDS1    240

#define BRIGHTNESS 255

//light settings 2
#define DATA_PIN2    4
#define CLK_PIN2     5
#define LED_TYPE2    APA102
#define COLOR_ORDER2 BGR
#define NUM_LEDS2    240

CRGB leds1[NUM_LEDS1];
CRGB leds2[NUM_LEDS2];

State ledState[] = {SOLID, SOLID};

//load cells
int buttonCutOff = 1000;

int b1_loadSample = 1;
const int b1_totalLoadCells = 1;

int b2_loadSample = 1;
const int b2_totalLoadCells = 1;

const byte b1_hx711_data_pin[] = {A0};
const byte b1_hx711_clock_pin[] = {A1};

const byte b2_hx711_data_pin[] = {A2};
const byte b2_hx711_clock_pin[] = {A3};

HX711 b1_loadCells[b1_totalLoadCells];
HX711 b2_loadCells[b2_totalLoadCells];

//event manager variables
int timer[] = { -1, -1};
CallBack timerCallback[] = {UNDEF, UNDEF};

//Color
CRGB ledColor[] = {CRGB::Black, CRGB::Black};

//FLASHCOLOR variables
int flashAmmount[] = {0, 0};
int timePerFlash[] = {0, 0};
int timeBetweenFlash[] = {0, 0};
bool isFlashLit[] = {false, false};

//RAINBOW variables
uint8_t hue[] = {0, 0};

//Countdown Variables
float countdownAmmount[] = {0, 0};
float PercentagePerTic[] = {0, 0};
int ticLength = 100;

//Time Since Start
unsigned long millisStart = 0;
unsigned int dt = 0;

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
    Rainbow(0);
  } else if (valBLUEGREEN == LOW) {
    deviceName = "BLUEGREEN";
    ChangeColor(CRGB::Blue, 0);
    ChangeColor(CRGB::Green, 1);
  } else if (valREDYELLOW == LOW) {
    deviceName = "REDYELLOW";
    ChangeColor(CRGB::Red, 0);
    ChangeColor(CRGB::Yellow, 1);
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

// Setup---Setup---Setup---Setup---Setup---Setup---Setup---Setup---Setup---Setup
void setup() {
  //init serial
  //setup arduino identification pins
  setupIDPin();

  //init serial
  Serial.begin (9600);
  Serial.setTimeout (50);
  inputString.reserve(200);

  //init name
  simonAssignDeviceName();

  //init leds
  FastLED.addLeds<LED_TYPE1, DATA_PIN1, CLK_PIN1, COLOR_ORDER1>(leds1, NUM_LEDS1).setCorrection(TypicalLEDStrip);
  FastLED.addLeds<LED_TYPE2, DATA_PIN2, CLK_PIN2, COLOR_ORDER2>(leds2, NUM_LEDS2).setCorrection(TypicalLEDStrip);
  delay(1000);
  FastLED.setBrightness(BRIGHTNESS);

  //Serial.println("Setting Solid Color");
  //ChangeColor(CRGB::Red, 1);
  //ChangeColor(CRGB::Green, 0);
  FastLED.show();

  //init load cells
  for (int i = 0; i < b1_totalLoadCells; i++)
  {
    b1_loadCells[i].begin(b1_hx711_data_pin[i], b1_hx711_clock_pin[i]);
  }
  for (int i = 0; i < b2_totalLoadCells; i++)
  {
    b2_loadCells[i].begin(b2_hx711_data_pin[i], b2_hx711_clock_pin[i]);
  }
  tare(1);
  tare(2);
  //Finished setup

}

// Loop---Loop---Loop---Loop---Loop---Loop---Loop---Loop---Loop---Loop---Loop
void loop() {
  //Serial.println(b1_loadCells[0].get_value());
  //serial
  if (stringComplete) {
    parseCommand();
  }
  //led
  dt = millis() - millisStart;
  millisStart = millis();
  updateEventManager(dt);
  updateLeds(dt);
  FastLED.show();
}


/*
  --------------------------------------------------------------------------------------------------------------------------
   This function flashes a color on the LEDs
   <color> is the color the light will flash
   <annount> is the ammount of times it will flash
   <timePerFlash> is the time the color is lit
   <timeBetweenFkash> is the time the color is off (if your only flashing once it dose not matter what you set this to)
   <ledStrip> is the stip it will flash on (0 or 1)
  --------------------------------------------------------------------------------------------------------------------------
*/
void FlashColor(CRGB color, int ammount, int _timePerFlash, int _timeBetweenFlash, int ledStrip)
{
  ledColor[ledStrip] = color;
  flashAmmount[ledStrip] = ammount;
  timePerFlash[ledStrip] = _timePerFlash;
  timeBetweenFlash[ledStrip] = _timeBetweenFlash;
  isFlashLit[ledStrip] = false;

  timerCallback[ledStrip] = FLASHCOLOR;
  timer[ledStrip] = 100;

  // Serial.println(timeBetweenFlash[ledStrip]);
}

/*
  --------------------------------------------------------------------------------------------------------------------------
   This function counts down
   <color> is the color the light
   <_time> is the ammount of time it will take
   <ledStrip> is the stip it will count down on
  --------------------------------------------------------------------------------------------------------------------------
*/
void Countdown(CRGB color, int _time, int ledStrip)
{
  ledState[ledStrip] = COUNTING;
  countdownAmmount[ledStrip] = 100;
  PercentagePerTic[ledStrip] = 100 / (_time / ticLength);
  ledColor[ledStrip] = color;
  ChangeColor(color, ledStrip);

  timer[ledStrip] = ticLength;
  timerCallback[ledStrip] = COUNTDOWN;
}

/*
  --------------------------------------------------------------------------------------------------------------------------
   This function makes the leds Rainbow
   <ledStrip> is the stip it will light on (0 or 1)
  --------------------------------------------------------------------------------------------------------------------------
*/
void Rainbow(int ledStrip)
{
  ledState[ledStrip] = RAINBOW;
}

/*
  --------------------------------------------------------------------------------------------------------------------------
   This function changes the color of the LEDs to a solid color
   <color> is the color the light will be
   <ledStrip> is the stip it will color (0 or 1)
  --------------------------------------------------------------------------------------------------------------------------
*/
void ChangeColor(CRGB color, int ledStrip)
{
  //ledState[ledStrip] = SOLID;

  if (ledStrip == 0)
  {
    fill_solid(leds1, NUM_LEDS1, color);
  }
  else
  {
    fill_solid(leds2, NUM_LEDS2, color);
  }

  //  FastLED.show();
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
    switch ( namePart[0]) {
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

    Countdown(_color,  colorTime.toInt(), 0);


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

void tare(int button)
{
  if (button == 1)
  {
    for (int i = 0; i < b1_totalLoadCells; i++)
    {
      b1_loadCells[i].tare();
    }
  } else if (button == 2)
  {
    for (int i = 0; i < b2_totalLoadCells; i++)
    {
      b2_loadCells[i].tare();
    }
  }
}

int getWeight(int button)
{
  int load = 0;
  int loadTally = 0;

  if (button == 1)
  {
    for (int i = 0; i < b1_totalLoadCells; i++)
    {
      int tl = 0;
      tl = b1_loadCells[i].get_value() / 100;
      if (tl > 0)
      {
        load += tl;
        loadTally++;
      }
      else
      {
        tl = 0;
      }

      //Serial.print("B" + String(i) + ": " + String(tl) + "   ");
    }
    //Serial.println(" ");
    if (loadTally == 0)
    {
      load = 0;
    }
    else
    {
      load = load / loadTally;
    }
  }
  else if (button == 2)
  {
    for (int i = 0; i < b2_totalLoadCells; i++)
    {
      int tl = 0;
      tl = b2_loadCells[i].get_value() / 100;
      if (tl > 0)
      {
        load += tl;
        loadTally++;
      }
      else
      {
        tl = 0;
      }

      //Serial.print("B" + String(i) + ": " + String(tl) + "   ");
    }
    //Serial.println(" ");
    if (loadTally == 0)
    {
      load = 0;
    }
    else
    {
      load = load / loadTally;
    }
  }
  return load;
}

//////////////////////////////////////////////////////////////////////////
//led stuff
//

void updateLeds(int deltaTime)
{
  updateRainbow(deltaTime);
  countdownUpdater(deltaTime);
}

void updateRainbow(int deltaTime)
{
  for (int i = 0; i < (sizeof(timer) / sizeof(int)); i++)
  {
    if (ledState[i] == RAINBOW)
    {
      EVERY_N_MILLISECONDS( 20 ) {
        hue[i]++;
      }
      switch (i)
      {
        case 0:
          fill_rainbow(leds1, NUM_LEDS1, hue[i], 2);
          break;

        case 1:
          fill_rainbow(leds2, NUM_LEDS2, hue[i], 2);
          break;
      }
      //FastLED.show();
    }
  }
}

void countdownUpdater(int deltaTime)
{
  for (int i = 0; i < (sizeof(timer) / sizeof(int)); i++)
  {
    if (ledState[i] == COUNTING)
    {
      EVERY_N_MILLISECONDS( 5 ) {
        switch (i)
        {
          case 0:
            fadeToBlackBy(leds1, NUM_LEDS1, 25);
            break;
          case 1:
            fadeToBlackBy(leds2, NUM_LEDS2, 25);
            break;
        }
      }
      int maxLed = 0;
      switch (i)
      {
        case 0:
          maxLed = NUM_LEDS1 - (((NUM_LEDS1 * 100) * (100 - countdownAmmount[0])) / 10000);
          //Serial.println(maxLed);
          for (int i2 = 0; i2 < maxLed; i2++)
          {
            leds1[i2] = ledColor[i];
          }

          break;
        case 1:
          maxLed = NUM_LEDS2 - (((NUM_LEDS2 * 100) * (100 - countdownAmmount[1])) / 10000);
          Serial.println(maxLed);
          for (int i2 = 0; i2 < maxLed; i2++)
          {
            leds2[i2] = ledColor[i];
          }
          break;
      }
    }
  }
}

//led updates
void flashColorUpdate(int ledStrip)
{
  if (flashAmmount[ledStrip] > 0)
  {
    if (isFlashLit[ledStrip])
    {
      ChangeColor(CRGB::Black, ledStrip);
      isFlashLit[ledStrip] = false;
      flashAmmount[ledStrip]--;
      timer[ledStrip] = timeBetweenFlash[ledStrip];
    }
    else
    {
      ChangeColor(ledColor[ledStrip], ledStrip);
      isFlashLit[ledStrip] = true;
      timer[ledStrip] = timePerFlash[ledStrip];

    }
  }
}

/*
   event manager functions
*/
void updateEventManager(int deltaTime)
{
  for (int i = 0; i < (sizeof(timer) / sizeof(int)); i++)
  {
    if (timer[i] > 0)
    {
      timer[i] -= deltaTime;
    }

    if (timer[i] <= 0 && timer[i] != -9999)
    {

      timer[i] = -9999;
      if (timerCallback[i] == FLASHCOLOR)
      {
        flashColorUpdate(i);
      }
      else if (timerCallback[i] == COUNTDOWN)
      {
        if (countdownAmmount[i] > 0)
        {
          countdownAmmount[i] -= PercentagePerTic[i];
          timer[i] = ticLength;
        }
      }
    }
  }
}
