var SerialPort = require('serialport');

////////////////

// the ports the arduinos are connected to
const SIMON_CENTER = 0;

const FIRST_BUTTON = 1; // set this to the index of the first arduino we consider a button
const SIMON_RED    = 1;
const SIMON_GREEN  = 2;
const SIMON_BLUE   = 3;
const SIMON_YELLOW = 4;
const LAST_BUTTON  = 4; // set this to the index of the last arduino we consider a button

var simonPorts = [5];

// this array will hold the latest weights from the buttons
var buttonWeights = [0, 0, 0, 0, 0]

// these are the ids of the arduinos that correspond to each controller
// This will let us correctly assing the right arduino to the right function
var pnpIds = {
  "USB\\VID_2341&PID_0042\\95635333331351416032" : SIMON_CENTER, 
  "USB\\VID_2341&PID_0042\\95635333331351801190" : SIMON_RED  

  // uncomment these when we have all buttons and set the correct strings
  //"USB\\VID_2341&PID_0042\\95635333331351801190" : SIMON_GREEN,  
  //"USB\\VID_2341&PID_0042\\95635333331351801190" : SIMON_BLUE,  
  //"USB\\VID_2341&PID_0042\\95635333331351801190" : SIMON_YELLOW  


}


////////////////////////////////////
// setup functions
//

// tell each of the connected ardies what they are and start their modes
function setupArduinos() {
      // setup parts
      simonPorts[SIMON_CENTER].write ('SIMON_CENTER' + '\n');
      simonPorts[SIMON_CENTER].write ('GS_ATTRACT' + '\n');

      simonPorts[SIMON_RED].write ('SIMON_RED' + '\n');
      // uncomment these lines when we have all four buttons
      //simonPorts[SIMON_GREEN].write ('SIMON_GREEN' + '\n');
      //simonPorts[SIMON_BLUE].write ('SIMON_BLUE' + '\n');
      //simonPorts[SIMON_YELLOW].write ('SIMON_RED' + '\n');

      readAllButtons(true);      
            
  
}



// tells the ports what functions should handle each event
function setupHandlers(port) {
    console.log ("setting up ports");

    port.on('open', showPortOpen);
    port.on('data', receiveSerialData);
    port.on('close', showPortClose);
    port.on('error', showError);

}


////////////////////////////////////
//
// Helper functions
//
//


function openPort (port) {
  port.open(function(error) {
      if (error) {
        console.log('failed to open port: ' + error);
      }
  });
}

function showPortOpen(path, options) {
  console.log('port open. Data rate: ' + path);
}

function showPortClose() {
   console.log('port closed.');
}
 
function showError(error) {
   console.log('Serial port error: ' + error);
}



function stripAlphaChars(source) {
  var out = source.replace(/[^0-9]/g, '');

  return out;
}


// the ardie sent us data
function receiveSerialData(data) {
  //console.log( String(data));
  //process.stdout.write(data);

  var ardId = -1;

   
  if (data.toString().indexOf("SIMON_CENTER") > -1) {
    ardId = SIMON_CENTER;
  } else if (data.toString().indexOf("SIMON_RED") > -1) {
    ardId = SIMON_RED;
  } else if (data.toString().indexOf("SIMON_GREEN") > -1) {
    ardId = SIMON_GREEN;
  } else if (data.toString().indexOf("SIMON_BLUE") > -1) {
    ardId = SIMON_BLUE;
  } else if (data.toString().indexOf("SIMON_YELLOW") > -1) {
    ardId = SIMON_YELLOW;
  }

  var bttnIdx = data.toString().indexOf("BTTN:");
  
  // put other data the button can send here
  //var fooIdx = data.toString().indexOf("FOO:")

  switch ( ardId ) {
    case SIMON_CENTER:
      break;

    case SIMON_RED:
    case SIMON_GREEN:
    case SIMON_BLUE:
    case SIMON_YELLOW:
      if (bttnIdx > -1){
        buttonWeights[ardId] = parseInt(data.toString().substring(bttnIdx+5),10);
      }
      break;
  }


}

// reads one button
function readButton(btnId) {
  simonPorts[btnId].write ('READ_BUTTONS' + '\n');
}

// reads all buttons
function readAllButtons(resetWeights) {

  if (resetWeights) {
    // clear everything
    for ( idx = FIRST_BUTTON; idx <= LAST_BUTTON; idx++ ) {
      buttonWeights[idx] = -1;
    }
  }

  simonPorts[SIMON_RED].write ('READ_BUTTONS' + '\n');
  //simonPorts[SIMON_GREEN].write ('READ_BUTTONS' + '\n');
  //simonPorts[SIMON_BLUE].write ('READ_BUTTONS' + '\n');
  //simonPorts[SIMON_YELLOW].write ('READ_BUTTONS' + '\n');

}


// returns the button with the most weight, or -1 if not all buttons have reported yet
function getPressedButton() {

  var idx = FIRST_BUTTON;
  var maxWeight = 0;
  var maxWeightIdx = idx;
  while (idx <= LAST_BUTTON) {
    if (buttonWeights[idx] == -1) {
      // not reported yet
      return -1;
    }
    if (buttonWeights[idx] >= maxWeight) {
      maxWeight = buttonWeights[idx];
      maxWeightIdx = idx;
    }
    idx = idx + 1;
  }

  console.log("button pushed is..." + maxWeightIdx.toString());


  return maxWeightIdx;

}


///////
// self-test method, will "push" the right button if ok is true, or the wrong one if false
//
function testFakeButtonPress(ok) {

  console.log("player sequence length " + playerSequence.length.toString())
  correctIdx = simonsSequence[playerSequence.length];

  for ( idx = FIRST_BUTTON; idx <= LAST_BUTTON; idx++ ) {
    wgt = 100
    if (idx == correctIdx) {
      if ( ok ) {
        wgt = 1000;
      }
      else {
        wgt = 10;;
      }
    }
    data = ""
    switch ( idx) {
      case SIMON_RED:
        data = "SIMON_RED : ";
        break;
      case SIMON_GREEN:
        data = "SIMON_GREEN : ";
        break;
      case SIMON_BLUE:
        data = "SIMON_BLUE : ";
        break;
      case SIMON_YELLOW:
        data = "SIMON_YELLOW : ";
        break;
    }
    data = data + "BTTN:" + wgt.toString();
    receiveSerialData(data)
  }


}

////////////////////////////////////
//
// Output logic
//
// These are the methods that control the lights and sound
//

// lights up the specified color and triggers the correct sound
function showColor(coloridx) {
  console.log("ShowColor: " + coloridx.toString());

  // send to raspberryPi
  // TBD

  // trigger audino
  // TBD


}


// showing the sequence Simon has picked for the player

// where we are in simon's sequence
var simonSequenceIdx = 0;
var simonIntervalTimer = null;

function showNextSimonColor() {
  if ( simonSequenceIdx < simonsSequence.length) {
    showColor(simonsSequence[simonSequenceIdx]);
    simonSequenceIdx++;
  }
  else {
    // we're done, start the player's timer
    if ( simonIntervalTimer ) {
      clearInterval(simonIntervalTimer);
    }
    gameState = GS_STARTTIMER;
  }
}


// playes the entire simon sequence
function showSimonsSequence() {
  gameState = GS_SHOWINGSEQ;  // set this so the game loop doesn't re-enter here
  console.log("Showing simon's sequence");
  simonSequenceIdx = 0;
  // how fast do we show simon?
  var simonms = 0;
  if (simonsSequence.length > 10 ) {
    simonms = 125;
  }
  else if (simonsSequence.length > 6 ) {
    simonms = 250;
  }
  else {
    simonms = 500;
  }

  simonIntervalTimer = setInterval (showNextSimonColor,simonms);
}


// starts the timer for the player to pick their button
function startPlayersTimer() {
  gameState = GS_TIMER;
  console.log("Starting player's timer");
  l = simonsSequence.lenth;

  timerms = 3000; // 3 seconds unless...
  if ( l > 15) {
    timerms = 1500; // gotta be quick!
  }

  // send message to center arduino
  // TBD

  // when time is up, read the buttons
  setTimeout(function(){
    gameState = GS_READBUTTONS;
  }, timerms);

}


////////////////////////////////////
//
// Game logic
//

// the sequence the computer has selected
var simonsSequence = []

// the sequence so far that the player has input
var playerSequence = []


function addNewColor() {
  console.log("Add new color");
  var coloridx = Math.floor(Math.random() * (LAST_BUTTON + 1 - FIRST_BUTTON) + FIRST_BUTTON);
  simonsSequence.push(coloridx);
  console.log(simonsSequence)
  gameState = GS_SHOWSEQ;   // now that we've added a color, show everyone the full sequence

}


// checks if the player's pattern so far matches
function checkPattern() {
  // note that player sequence will always be less than simon's sequence
  for ( idx = 0; idx < playerSequence.length; idx++  ) {
    if ( playerSequence[idx] != simonsSequence[idx]) {
      return false;
    } 
  }
  return true;

}


// checks which button is pressed and uses that as the player's choice
function makePlayersChoice() {
  var selected = getPressedButton();
  if ( selected > -1 ) {
    gameState = GS_EVALUATING;  // set this so the loop doesn't re-enter this function
    console.log("Make player choice");
    playerSequence.push(selected);

    if ( checkPattern() ) {
      // success, show pushed color
      showColor(selected);
      if ( playerSequence.length == simonsSequence.length ) {
        // they got it!   computer's turn
        console.log("++++ player got entire sequence");
        playerSequence = [];
        gameState = GS_COMPUTER;  // computer's turn
      }
      else {
        // so far so good, but they haven't finished the sequence yet, start timer for next button
        gameState = GS_STARTTIMER;
      }
    

    }
    else {
      // bzzzzzz - failed, game over
      console.log("GAME OVER!!!!");

      gameState = GS_GAME_OVER;

    }

  }

  
}


// start a brand new game
function newGame() {
  console.log("New Game!");
  simonsSequence = [];
  playerSequence = [];
  gameState = GS_COMPUTER;
}






////////////////////////////////////
//
// Main functions and loop
//
//

// The first thing we do is find and connect all the arduinos connected as USB/serial ports 
SerialPort.list(function (err, ports) {
  ports.forEach(function(port) {
    console.log(port.comName);
    console.log(port.pnpId);
    console.log(port.manufacturer);
    var newPort = new SerialPort (port.comName,{
      baudrate: 115200,
      autoOpen: false
    });
    if (port.pnpId in pnpIds ) {
      console.log("assinging " + port.pnpId + " to " + (pnpIds[port.pnpId]).toString() )
      simonPorts[pnpIds[port.pnpId]] = newPort;
    } 
    else {
      console.log ("unknown port " + port.pnpId);
    }
    setupHandlers(newPort);
    openPort (newPort);
  });
});


/*
 * Main Simon Game State Loop
 */

// GAME STATES
const GS_INIT       = -1; // waiting to set up
const GS_ATTRACT    = 1;  // running attract mode, waiting for RED to be pressed to start the game
const GS_COMPUTER   = 2;  // Player got sequence right, Computer is thinking, adding a new color to the sequence
const GS_SHOWSEQ    = 3;  // Start the process of showing the computer's sequence
const GS_SHOWINGSEQ = 4;  // we are showing the sequence.  Just wait...
const GS_STARTTIMER = 5;  // Start the timer for the player to push the next button
const GS_TIMER      = 6;  // timer is running down, just wait...
const GS_READBUTTONS= 7;  // reading the button weights
const GS_PLAYER     = 8;  // timer ran our, figure out what button the player pushed
const GS_EVALUATING = 9;  // evaluate if the player pushed the right button
const GS_GAME_OVER  = 10; // player got the sequence wrong...

var gameState = GS_INIT;



// this will start the loop 5 seconds after we start up (to allow ports to be initialized), then
// start a 50 ms loop
setTimeout(function(){
  console.log("Starting Loop")
  setInterval (loop,50);
}, 5000);


// the main game loop
function loop () {

  switch( gameState) {
    case GS_INIT:
      setupArduinos();
      gameState = GS_ATTRACT;
      break;

    case GS_ATTRACT:
      readButton(SIMON_RED)
      if (buttonWeights[SIMON_RED] > 0) {
        console.log ("Button Pushed :: Start Game");
        simonPorts[SIMON_CENTER].write ('GS_COMPUTER' + '\n');
        newGame();
      }
      ////////////
      // enable this to auto-test the logic
      newGame();
      //
      ///////////
      break;

    case GS_COMPUTER:
      // simon is taking his turn. 
      addNewColor();
      break;

    case GS_SHOWSEQ:
      showSimonsSequence();
      break;

    case GS_SHOWINGSEQ:
      // be patient while simon shows the player the sequence
      break;

    case GS_STARTTIMER:
      // put the center into countdown mode
      startPlayersTimer();
      break;

    case GS_TIMER:
      break;

    case GS_READBUTTONS:
      // timer ran out, check the buttons to see what the player has pressed 
      readAllButtons(true);
      // uncomment to run self-test sequence
      testFakeButtonPress(playerSequence.length < 20);
      
      gameState = GS_PLAYER;
      break;
      
    case GS_PLAYER:
      makePlayersChoice();
      break;

    case GS_EVALUATING:
      // checking to see if the player got it right...
      break;

    case GS_GAME_OVER:
      break;
  }


}
  