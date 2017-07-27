#include <FastLED.h>

#define DATA_PIN    2
#define CLK_PIN     3
#define LED_TYPE    APA102
#define COLOR_ORDER BGR
#define NUM_LEDS    144

#define BRIGHTNESS 15

CRGB leds[NUM_LEDS];
int readingPin = 0;
int reading;


#define GAME_STATE_ATTRACT  0
#define GAME_STATE_COMPUTER 1
#define GAME_STATE_PLAYER   2
#define GAME_STATE_ENDGAME  3
#define GAME_STATE_READ_BUTTONS  4
int gameState = -1;
int gameTimer = 2;
boolean timerCompleted = false;
boolean buttonIsPushed = false;
boolean playerCorrect = false;
int buttonPushed = -1;

int simonPattern[2048]; 
long simonPatternLength = 0;
int playerIndex = 0;

int timerDot = 0;


#define DEBOUNCE 10  // button debouncer, how many ms to debounce, 5+ ms is usually plenty

// here is where we define the buttons that we'll use. button "1" is the first, button "6" is the 6th, etc
byte buttons[] = {4,5,6,7}; // the analog 0-5 pins are also known as 14-19
char* buttonName[] = {"Red", "Blue", "Green", "Yellow"};
// This handy macro lets us determine how big the array up above is, by checking the size
#define NUMBUTTONS sizeof(buttons)
// we will track if a button is just pressed, just released, or 'currently pressed' 
byte pressed[NUMBUTTONS], justpressed[NUMBUTTONS], justreleased[NUMBUTTONS];

uint8_t gHue = 0;

String inputString = "";
boolean stringComplete = false; 

// defines for serial communication to devices
String deviceName = "unknown";


void setup() {
  // put your setup code here, to run once:
  Serial.begin (115200);
  Serial.setTimeout (50);
  FastLED.addLeds<LED_TYPE, DATA_PIN, CLK_PIN, COLOR_ORDER>(leds, NUM_LEDS).setCorrection(TypicalLEDStrip);
  delay(1000);
  FastLED.setBrightness(BRIGHTNESS);
  FastLED.show();
  //glowAllLEDS ();
  //gameState = GAME_STATE_ATTRACT;

  inputString.reserve(200);

  for (int index = 0; index < NUMBUTTONS; index++) {
    pinMode(buttons[index], INPUT_PULLUP);
  }
  
}

void loop() {
  // put your main code here, to run repeatedly:

  //Serial.println (gameState); 

  if (gameState == GAME_STATE_ATTRACT){
    //Serial.println ("play attract loop");  
    simonAttractLoop();
     
  } 
  
  // print the string when a newline arrives:
  if (stringComplete) {

    parseCommand();

  }
  
  FastLED.show(); 
}


void simonOutput (String outputString) {
    Serial.print(deviceName);
    Serial.print(" : ");
    Serial.println(outputString);         
}

void parseCommand () {


    if (inputString == "SIMON_CENTER") {
      deviceName = "SIMON_CENTER";
      Serial.print("DEVICE ASSIGNED : ");
      Serial.println(deviceName);       
    } else if (inputString == "SIMON_RED") {
      deviceName = "SIMON_RED";
      Serial.print("DEVICE ASSIGNED : ");
      Serial.println(deviceName);       
    }

    if (inputString == "GS_ATTRACT") {
      gameState = GAME_STATE_ATTRACT;
      simonOutput(inputString);       
    } else if (inputString == "GS_COMPUTER") {
      gameState = GAME_STATE_COMPUTER;
      simonComputerTurn();       
    } else if (inputString == "READ_BUTTONS") {
      //simonOutput(inputString);       
      simonReadButtonWeight();
    }
    
    inputString = "";
    stringComplete = false;
  
}

/*
 * Simon main states
 */


void simonAttractLoop() {
  attractLoop();
}


void simonReadButtonWeight() {
  readPushButtons();
  if (buttonIsPushed) {
    simonOutput("BTTN:" + String(random(1000, 2000)));
    buttonIsPushed = false;
  } else {
    simonOutput("BTTN:0");
  }
}

void simonComputerTurn() {

  gameState = GAME_STATE_COMPUTER;
  playerIndex = 0;
  fadeOutLEDS();

  
  // create new color and add it to array
  long randColor;
  randomSeed(analogRead(0));
  randColor = random(4);
  Serial.println(String(randColor)); 
  
  simonPattern[simonPatternLength] = randColor;
  simonPatternLength++;
  
  // display colors in array
  DisplaySimonPattern();
  gameState = GAME_STATE_PLAYER;
}

void simonPlayerTurn() {

     buttonPushed = -1;
  
    if ( !timerCompleted) {
      fastLoopLED();    
    } else {
      //readScaleButton();
      readPushButtons();  

      if (buttonIsPushed) {
        glowAllLEDS(buttonName[buttonPushed]); 
        FastLED.show(); 
        delay(2000);

        fill_solid( leds, NUM_LEDS, CRGB::Black );
        FastLED.show();
        delay(1000);
      }

      checkPattern();

      buttonIsPushed = false;
      timerCompleted = false; 

      
  
    }
  
}

void simonEndGame () {
  blinkLEDRed();
  gameState = GAME_STATE_ATTRACT;
}


void checkPattern() {
  if (simonPattern[playerIndex] == buttonPushed) {
    Serial.println("Correct");
    playerCorrect = true;

    Serial.println(playerIndex);
    Serial.println(simonPatternLength);

    if (playerIndex == simonPatternLength-1) {
      gameState = GAME_STATE_COMPUTER;
    } else {
      playerIndex++;
    }
    
  } else {
     Serial.println("BAD!");
    playerCorrect = false;
    gameState = GAME_STATE_ENDGAME;
  }
}

void DisplaySimonPattern () {

  int pos;
  for (int j; j< 1000; j++) {
    // a colored dot sweeping back and forth, with fading trails
    fadeToBlackBy( leds, NUM_LEDS, 20);
    pos = beatsin16(13,0,NUM_LEDS);
    leds[pos] += CHSV( 30, 255, 192);
    FastLED.show();
  }

  for (int j=0; j<100; j++) { 
    fadeToBlackBy( leds, NUM_LEDS, 20);
    FastLED.show();
  }
    
  delay (250);  
  
  for (int i; i < simonPatternLength; i++) {
    if (simonPattern[i] == 0) {
      fill_solid( leds, NUM_LEDS, CRGB::Red );
    } else if (simonPattern[i] == 1) {
      fill_solid( leds, NUM_LEDS, CRGB::Blue );
    } else if (simonPattern[i] == 2) {
      fill_solid( leds, NUM_LEDS, CRGB::Green );
    } else if (simonPattern[i] == 3) {
      fill_solid( leds, NUM_LEDS, CRGB::Yellow );
    }
    FastLED.show();
    delay(500);
    fill_solid( leds, NUM_LEDS, CRGB::Black );
    FastLED.show();
    delay(250);
    
  }
  delay (500);
}

/*
 * Button functions
 */

void readPushButtons() {
  static byte currentstate[NUMBUTTONS];

  byte index = 0;

  for (index = 0; index < NUMBUTTONS; index++) {
    currentstate[index] = digitalRead(buttons[index]);   // read the button
        
    //Serial.print(currentstate[index], DEC);    
    if (currentstate[index] == 0 && !buttonIsPushed) {
      buttonIsPushed = true;
      buttonPushed = index;
      
        
    }
  }
      

}


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
 * LED functions
 */

void fadeOutLEDS() {
  for (int i; i < 30; i++) {
    fadeToBlackBy( leds, NUM_LEDS, 30);
    FastLED.show();
    delay (50);
  }
}

void attractLoop() 
{
  EVERY_N_MILLISECONDS( 20 ) { gHue++; }
  fill_rainbow( leds, NUM_LEDS,gHue , 7);
  
//  if( random8() < 80) {
//    leds[ random16(NUM_LEDS) ] += CRGB::White;
//  }
  
}


void blinkLEDRed () {
  for( int i = 0; i < 10; i++) { 
      fill_solid( leds, NUM_LEDS, CRGB::Red );
      FastLED.show();
      delay (100);    
      fill_solid( leds, NUM_LEDS, CRGB::Black );
      FastLED.show();
      delay (100);
  }
}

void glowAllLEDS (String colorChoice) {
  
  //Serial.println (colorChoice);

  
    if (colorChoice == "Blue") {
 
      fill_solid( leds, NUM_LEDS, CRGB::Blue );
    } else if (colorChoice == "Red") {

      fill_solid( leds, NUM_LEDS, CRGB::Red );
    } else if (colorChoice == "Green") {
   
      fill_solid( leds, NUM_LEDS, CRGB::Green );   
    } else  {
    
      fill_solid( leds, NUM_LEDS, CRGB::Yellow );
    }
    
}


void fastLoopLED () {
          //for(int timerDot = 0; timerDot <= NUM_LEDS; timerDot++) { 

  if (buttonIsPushed) {
    return; 
  }

  for( int i = timerDot; i < NUM_LEDS; i++) { 
      leds[i] = CRGB::White;    
  }

  
  //leds[timerDot] = CRGB::Blue;
  fadeToBlackBy( leds, NUM_LEDS, 40);
  //FastLED.show();
  // clear this led for the next time around the loop
  //leds[timerDot] = CRGB::Black;
  delay((1000 / NUM_LEDS * gameTimer));

  if (timerDot >= NUM_LEDS+25) {
    timerDot = 0;
    timerCompleted = true;            
  } else {
    timerDot++;
  }
      
}
