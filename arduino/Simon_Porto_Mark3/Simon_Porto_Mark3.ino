#include <Thread.h>
#include <ThreadController.h>
#include <FastLED.h>
#include "HX711.h"

//load cells
int loadSample = 5;
const int totalLoadCells = 4;

const byte hx711_data_pin[] = {A2, A4, A6, A8};
const byte hx711_clock_pin[] = {A3, A5, A7, A9};

HX711 loadCells[totalLoadCells];


//light settings
#define DATA_PIN    2
#define CLK_PIN     3
#define LED_TYPE    APA102
#define COLOR_ORDER BGR
#define NUM_LEDS    144
#define BRIGHTNESS 15

CRGB leds[NUM_LEDS];

//settup threads
ThreadController tController = ThreadController();

Thread colorPatternThread = Thread();

//enum settup
enum state {
  ATTRACT,
  COLORPATTERN,
  COLOR,
  WEIGHTCOLOR
};


//init global vars
state gameState = ATTRACT;


//init attract vars
uint8_t gHue = 0;


//init colorPattern vars
int lightDelay;
CRGB colorPattern[0];
int colorPatternLength = 0;
int colorIndex = 0;
boolean patternRunning = false;

//init color vars
CRGB solidColor = CRGB::Red;

//init weightColor vals
int weightModifier = 1;

void setup() {
  Serial.begin(115200);

  //init threads
  colorPatternThread.onRun(colorPatternCallback);
  colorPatternThread.enabled = false;
  tController.add(&colorPatternThread);


  //init load cells
  for (int i = 0; i < totalLoadCells; i++)
  {
    loadCells[i].begin(hx711_data_pin[i], hx711_clock_pin[i]);
  }
  tar();


  //init leds
  FastLED.addLeds<LED_TYPE, DATA_PIN, CLK_PIN, COLOR_ORDER>(leds, NUM_LEDS).setCorrection(TypicalLEDStrip);
  delay(1000);
  FastLED.setBrightness(BRIGHTNESS);
  FastLED.show();
}

void loop() {
  tController.run();
  //run states
  switch (gameState)
  {
    case ATTRACT:
      updateAttract();
      break;

    case COLORPATTERN:
      updateColorPattern();
      break;

    case COLOR:
      updateColor();
      break;
  }

  FastLED.show();
}

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

void displayColorPattern() {
  patternRunning = true;
  colorIndex = 0;
  colorPatternThread.enabled = true;
}

void updateAttract() {
  EVERY_N_MILLISECONDS( 20 ) {
    gHue++;
  }
  fill_rainbow( leds, NUM_LEDS, gHue , 7);
}

void updateColorPattern() {
  fill_solid(leds, NUM_LEDS, colorPattern[colorIndex]);
}

void colorPatternCallback() {
  colorIndex++;
  if (colorIndex == colorPatternLength)
  {
    patternRunning = false;
    colorPatternThread.enabled = false;
    fill_solid(leds, NUM_LEDS, CRGB::Black);
  }
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

