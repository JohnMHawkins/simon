#include <TimerThree.h>
#include <FastLED.h>
#include "HX711.h"

//enum settup
enum State {
  GS_INIT,
  GS_ATTRACT,
  GS_SHOWINGSEQ,
  GS_EVALUATING,
  GS_GAME_OVER
};

enum LedState {
  SOLID,
  WEIGHT,
  PATTERN,
  RAINBOW,
  COUNTDOWN,
  FLASH,
  STREAK
};


//load cells
int b1_loadSample = 5;
const int b1_totalLoadCells = 1;

int b2_loadSample = 5;
const int b2_totalLoadCells = 1;

const byte b1_hx711_data_pin[] = {A0};
const byte b1_hx711_clock_pin[] = {A2};

const byte b2_hx711_data_pin[] = {A1};
const byte b2_hx711_clock_pin[] = {A3};

HX711 b1_loadCells[b1_totalLoadCells];
HX711 b2_loadCells[b2_totalLoadCells];


//light settings
#define DATA_PIN    2
#define CLK_PIN     3
#define LED_TYPE    APA102
#define COLOR_ORDER BGR
#define NUM_LEDS    144
#define BRIGHTNESS 10

CRGB leds[NUM_LEDS];

// vars for serial communication
// vars for serial communication
String inputString = "";
String simonCommand = "";
boolean stringComplete = false;
String deviceName = "unknown";
String deviceInput = "unknown";

//Arduino Identification Pins (set to ground)
int PIN_CENTER = 51;

int PIN_BLUEGREEN = 52;
int PIN_REDYELLOW = 53;

int PIN_BLUE = 50;
int PIN_RED = 49;
int PIN_GREEN = 48;
int PIN_YELLOW = 47;

//settup threads
//ThreadController tController = ThreadController();

//Thread colorPatternThread = Thread();

//init global vars
State gameState = GS_INIT;
LedState ledState = RAINBOW;


//init attract vars
uint8_t gHue = 0;

//lightStreak vars
int streakAmmount;
int streakTime;
int streakFade;
int streakFadeAmmount;

#define maxStreakAmmount 99

CRGB streaks[maxStreakAmmount];
int streakPos[maxStreakAmmount];
int streakDirections[maxStreakAmmount];

int streakColorAmmount = 3;
CRGB streakColors[] = {CRGB::Red, CRGB::Green, CRGB::Blue};



//init colorPattern vars
int lightDelay;
CRGB colorPattern[1];
int colorPatternLength = 0;
int colorIndex = 0;
int colorTimer = 0;
boolean patternRunning = false;

//init color vars
CRGB solidColor = CRGB::Red;

//countdown vars
int tickNumber = 144;
int tickTime;
int totalTime;
int trailLength = 19;
int countdownIndex = 0;

//flashLed vars
int flashCount;
int flashTime;
int flashIndex = 0;
CRGB flashColor;

//init weightColor vals
int weightModifier = 1;


void initGame()
{
  solidColor = CRGB::DodgerBlue;
  ledState = SOLID;
}

void attract()
{
  ledState = RAINBOW;
}

void showingSeq()
{
  ledState = PATTERN;
}

void player()
{
  solidColor = CRGB::Red; //change to color of button with a var later
  ledState = WEIGHT;
}

void gameOver()
{
  solidColor = CRGB::DarkRed;
  ledState = SOLID;
}

void setup() {
  Timer3.initialize(1000000);

  randomSeed(analogRead(0));
  //setup arduino identification pins
  setupIDPin();

  //init serial
  Serial.begin (115200);
  Serial.setTimeout (50);
  inputString.reserve(200);

  //init threads
  //  colorPatternThread.onRun(colorPatternCallback);
  //  colorPatternThread.enabled = false;
  //  tController.add(&colorPatternThread);

  //init name
  simonAssignDeviceName();
  //init load cells
  for (int i = 0; i < b1_totalLoadCells; i++)
  {
    b1_loadCells[i].begin(b1_hx711_data_pin[i], b1_hx711_clock_pin[i]);
  }
  for (int i = 0; i < b2_totalLoadCells; i++)
  {
    b2_loadCells[i].begin(b2_hx711_data_pin[i], b2_hx711_clock_pin[i]);
  }
  tar(1);
  tar(2);
  //loadCell.begin(A0, A1);


  //init leds
  FastLED.addLeds<LED_TYPE, DATA_PIN, CLK_PIN, COLOR_ORDER>(leds, NUM_LEDS).setCorrection(TypicalLEDStrip);
  delay(1000);
  FastLED.setBrightness(BRIGHTNESS);
  FastLED.show();
}

int getLoadCellPin(String input)
{
  if (deviceName == "REDYELLOW")
  {
    if (input == "RED")
    {
      return 1;
    }
    else if (input == "YELLOW")
    {
      return 2;
    }
  }
  else if (deviceName == "BLUEGREEN")
  {
    if (input == "BLUE")
    {
      return 1;
    }
    else if (input == "GREEN")
    {
      return 2;
    }
  }
  return 1;
}

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

void loop() {
  //  tController.run();
  //Serial.println(getWeight(1));

  if (colorTimer > 0)
  {
    colorTimer --;
  }

  //run led states
  switch (ledState)
  {
    case RAINBOW:
      updateAttract();
      break;

    case PATTERN:
      updateColorPattern();
      break;

    case SOLID:
      updateColor();
      break;

    case WEIGHT:
      weightColor(1);
      break;

    case COUNTDOWN:
      countdownLeds();
      break;

    case FLASH:
      break;

    case STREAK:
      break;
  }

  noInterrupts();
  if (stringComplete) {
    parseCommand();
  }
  interrupts();

  FastLED.show();
}

//SERIAL STUFF------------------------------------------
//
// helper function to output to serial port
//
void simonOutput (String outputString) {
  Serial.print(deviceName);
  Serial.print("[" + deviceInput + "]");
  Serial.print(":");
  Serial.println(outputString);
}

void parseCommand () {

  //our giant if statement to compare all commands

  //assign device nmae
  if (inputString.substring(0, 4) == "NAME" ) {
    //assign device nmae
    simonAssignDeviceName();

  } else {
    //get deviceInputName and process command
    deviceInput = getValue(inputString, ':', 0);
    simonCommand = getValue(inputString, ':', 1);
    //Serial.println(deviceInput + "  " + simonCommand);

    processSimonCommand();
  }

  // clear inputString, deviceInput & commandString
  inputString = "";
  deviceInput = "";
  simonCommand = "";
  stringComplete = false;

}

void processSimonCommand() {

  //  if (inputString.substring(0,10) == "FLASHCOLOR" ) {
  if (simonCommand == "FLASHCOLOR") {
    //simonOutput(inputString);
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

    flashColors(_color, flashTime.toInt(), flashAmmount.toInt());

  } else if (simonCommand == "GS_ATTRACT") {
    simonOutput (simonCommand);
    attract();
    //Call your attract function here

  } else if (simonCommand == "GS_COMPUTER") {
    simonOutput (simonCommand);
    //setSolidColor(CRGB::Orange);
    lightStreak(1, 1000, 50);
    //Call compiter function here

  } else if (simonCommand == "GS_TIMER") {
    simonOutput (simonCommand);
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

    countdown(_color,  colorTime.toInt());

  } else if (simonCommand == "READ_BUTTONS") {
    //create test button data, replace with real button function
    int tmpDeviceNum = getLoadCellPin(deviceInput);
    if (tmpDeviceNum == 1)
    {
      simonOutput ("BTTN:" + String(getWeight(1)));
    }
    else if (tmpDeviceNum == 2)
    {
      simonOutput ("BTTN:" + String(getWeight(2)));
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


/*
   all functions related to parse commands

*/



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

//NOTSERIAL STUFF----------------------------------------------

void ledRainbow()
{
  ledState = RAINBOW;
}

void ledPattern(int interval) {
  colorPatternLength = 1;
  ledState = PATTERN;
  patternRunning = true;
  colorIndex = 0;
  colorTimer = interval;
  unsigned long _interval = 1000l * interval;
  Timer3.setPeriod(_interval);
  //Serial.println(_interval);
  Timer3.restart();
  Timer3.attachInterrupt(colorPatternCallback);//interval * 1000);
  //colorPatternThread.setInterval(interval);
  //colorPatternThread.enabled = true;
}

//void ledSolid(CRGB)

void tar(int button)
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

  if (button == 1)
  {
    for (int i = 0; i < b1_totalLoadCells; i++)
    {
      load += b1_loadCells[i].get_value() / 10;
    }

    load = load / b1_totalLoadCells;
  }
  else if (button == 2)
  {
    for (int i = 0; i < b2_totalLoadCells; i++)
    {
      load += b2_loadCells[i].get_value() / 10;
    }

    load = load / b2_totalLoadCells;

  }
  if (load < 0)
  {
    load = 0;
  }
  return load;
}

void updateAttract() {
  EVERY_N_MILLISECONDS( 20 ) {
    gHue++;
  }
  fill_rainbow( leds, NUM_LEDS, gHue , 7);
}

void updateColorPattern() {
  if (patternRunning == true)
  {
    fill_solid(leds, NUM_LEDS, colorPattern[0]);
  }
  else
  {
    //Timer3.detachInterrupt();
  }
}

void colorPatternCallback() {
  if (colorIndex > 0)
  {
    fill_solid(leds, NUM_LEDS, CRGB::Black);
    patternRunning = false;
    colorIndex = 0;
    Timer3.detachInterrupt();
  }
  else
  {
    colorIndex++;
  }
  //  colorPatternThread.enabled = false;
}

void setSolidColor(CRGB _color)
{
  solidColor = _color;
  ledState = SOLID;
}

void countdownLeds()
{

}

void flashColors(CRGB _color, int _time, int _ammount)
{
  flashCount = _ammount;
  flashTime = _time;
  flashIndex = 0;
  flashColor = _color;
  ledState = FLASH;

  unsigned long _interval = 1000l * (flashTime / 2);
  Timer3.setPeriod(_interval);
  Timer3.restart();
  Timer3.attachInterrupt(flashColorsCallback);
}

void flashColorsCallback()
{
  if (flashIndex % 2 == 1)
  {
    fill_solid(leds, NUM_LEDS, CRGB::Black);
  }
  else
  {
    fill_solid(leds, NUM_LEDS, flashColor);
  }
  flashIndex ++;
  if (flashIndex > flashCount * 2)
  {
    fill_solid(leds, NUM_LEDS, CRGB::Black);
    Timer3.detachInterrupt();
  }
}

void lightStreak (int _ammount, int _time, int fade)
{
  streakAmmount = _ammount;
  streakFadeAmmount = fade;
  ledState = STREAK;
  fill_solid(leds, NUM_LEDS, CRGB::Black);
  FastLED.show();

  for (int i = 0; i < streakAmmount; i++)
  {
    streakPos[i] = random(0, NUM_LEDS);
    if (random(0, 2) == 1)
    {
      streakDirections[i] = 1;
    }
    else
    {
      streakDirections[i] = -1;
    }

    streaks[i] = streakColors[random(0, streakColorAmmount)];
  }

  unsigned long _interval = 1000l * (streakTime / NUM_LEDS);
  Timer3.setPeriod(_interval);
  Timer3.restart();
  Timer3.attachInterrupt(lightStreakCallback);

}

void lightStreakCallback()
{
  for (int i = 0; i < streakAmmount; i++)
  {
    leds[streakPos[i]] += streaks[i];
    streakPos[i] += streakDirections[i];
    if (streakPos[i] < 0)
    {
      streakPos[i] = NUM_LEDS - 1;
    }
    if (streakPos[i] >= NUM_LEDS)
    {
      streakPos[i] = 0;
    }
  }

  if (ledState != STREAK)
  {
    fill_solid(leds, NUM_LEDS, CRGB::Black);
    Timer3.detachInterrupt();
  }
  fadeToBlackBy(leds, NUM_LEDS, 50);
}


void countdown(CRGB _color, int _time)
{
  solidColor = _color;
  totalTime = _time;
  ledState = COUNTDOWN;
  tickTime = totalTime / (tickNumber + trailLength);
  countdownIndex = 0;
  unsigned long _interval = 1000l * tickTime;
  Timer3.setPeriod(_interval);
  Timer3.restart();
  Timer3.attachInterrupt(countdownCallback);
}

void countdownCallback () {

  fadeToBlackBy( leds, NUM_LEDS, 64);

  int cIndex = countdownIndex;
  if (cIndex > tickNumber)
  {
    cIndex = tickNumber;
  }

  for (int i = 0; i < tickNumber - cIndex; i++)
  {
    leds[i] = solidColor;
  }

  countdownIndex++;
  if (countdownIndex > tickNumber + trailLength)
  {
    countdownIndex = 0;
    fill_solid(leds, NUM_LEDS, CRGB::Black);
    Serial.println("Timer Done");
    Timer3.detachInterrupt();
  }
}

void setColorFlash(CRGB _color, int _time)
{
  CRGB _colorArray[] = {_color};
  colorPattern[0] = _colorArray[0];
  ledPattern(_time);
}

void updateColor() {
  fill_solid( leds, NUM_LEDS, solidColor);
}

void weightColor(int button) {
  fill_solid( leds, NUM_LEDS, solidColor);
  int brightness;

  if (button == 1)
  {
    brightness = getWeight(1) / weightModifier;
  }
  else if (button == 2)
  {
    brightness = getWeight(2) / weightModifier;
  }
  if (brightness > 255)
  {
    brightness = 255;
  }

  FastLED.setBrightness(brightness);
}
