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
  COUNTDOWN
};


//load cells
int loadSample = 5;
const int totalLoadCells = 1;

const byte hx711_data_pin[] = {A1, A4, A6, A8};
const byte hx711_clock_pin[] = {A0, A5, A7, A9};

HX711 loadCells[totalLoadCells];
HX711 loadCell;


//light settings
#define DATA_PIN    2
#define CLK_PIN     3
#define LED_TYPE    APA102
#define COLOR_ORDER BGR
#define NUM_LEDS    144
#define BRIGHTNESS 10

CRGB leds[NUM_LEDS];

// vars for serial communication
String inputString = "";
boolean stringComplete = false;
String deviceName = "unknown";

//settup threads
//ThreadController tController = ThreadController();

//Thread colorPatternThread = Thread();

//init global vars
State gameState = GS_INIT;
LedState ledState = RAINBOW;


//init attract vars
uint8_t gHue = 0;


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

  //init serial
  Serial.begin (115200);
  Serial.setTimeout (50);
  inputString.reserve(200);

  //init threads
  //  colorPatternThread.onRun(colorPatternCallback);
  //  colorPatternThread.enabled = false;
  //  tController.add(&colorPatternThread);


  //init load cells
  for (int i = 0; i < totalLoadCells; i++)
  {
    loadCells[i].begin(hx711_clock_pin[i], hx711_data_pin[i]);
  }
  tar();
  loadCell.begin(A0, A1);


  //init leds
  FastLED.addLeds<LED_TYPE, DATA_PIN, CLK_PIN, COLOR_ORDER>(leds, NUM_LEDS).setCorrection(TypicalLEDStrip);
  delay(1000);
  FastLED.setBrightness(BRIGHTNESS);
  FastLED.show();
}

void loop() {
  //  tController.run();
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
      weightColor();
      break;

    case COUNTDOWN:
      countdownLeds();
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
  Serial.print(":");
  Serial.println(outputString);
}

//
// function to compare our inputString to possible commands
// add more commands and calls to funcitons here
//
void parseCommand () {

  //our giant if statement to compare all commands

  if (inputString.substring(0, 4) == "NAME" ) {
    simonAssignDeviceName();

  } else if (inputString.substring(0, 10) == "FLASHCOLOR" ) {
    //simonOutput(inputString);
    String colorRGB = getValue(inputString, ':', 1);

    String colorRed = getValue(colorRGB, '-', 0);
    String colorGreen = getValue(colorRGB, '-', 1);
    String colorBlue = getValue(colorRGB, '-', 2);

    String colorTime = getValue(inputString, ':', 2);

    CRGB _color;
    _color.r = colorRed.toInt();
    _color.g = colorGreen.toInt();
    _color.b = colorBlue.toInt();

    setColorFlash(_color, colorTime.toInt());
    simonOutput("COLOR:" + colorRed + "::" + colorGreen + "::" + colorBlue + " TIME:" + colorTime);

  } else if (inputString == "GS_ATTRACT") {
    simonOutput (inputString);
    attract();
    //Call your attract function here

  } else if (inputString == "GS_COMPUTER") {
    simonOutput (inputString);
    setSolidColor(CRGB::Orange);
    //Call compiter function here

  } else if (inputString == "GS_TIMER") {
    simonOutput (inputString);
    //Call your timer function here
    countdown(CRGB::White,  1000G);

  } else if (inputString == "READ_BUTTONS") {
    //create test button data, replace with real button function
    //Serial.println(loadCell.read());
    simonOutput ("BTTN:" + String(getWeight()));

  } else {
    simonOutput ("NOCMD:" + inputString);
  }

  // clear inputString
  inputString = "";
  stringComplete = false;

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

  deviceName = getValue(inputString, ':', 1);

  Serial.print("DEVICE ASSIGNED : ");
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

void tar()
{
  for (int i = 0; i < totalLoadCells; i++)
  {
    loadCells[i].tare();
  }
}

int getWeight()
{
  int load = 0;

  for (int i = 0; i < totalLoadCells; i++)
  {
    load += loadCells[i].get_value(loadSample) / 10;
  }

  load = load / totalLoadCells;

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
  if(cIndex > tickNumber)
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

void weightColor() {
  fill_solid( leds, NUM_LEDS, solidColor);

  int brightness = getWeight() / weightModifier;

  if (brightness > 255)
  {
    brightness = 255;
  }

  FastLED.setBrightness(brightness);
}
