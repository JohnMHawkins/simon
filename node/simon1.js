var SerialPort = require('serialport');


// global vars

//var simonRedBttnWeight = 0;




// the ports the arduinos are connected to

const SIMON_CENTER = 0;
const SIMON_RED    = 1;
const SIMON_GREEN  = 2;
const SIMON_BLUE   = 3;
const SIMON_YELLOW = 4;

var simonPorts = [5];


var buttonWeights = [0, 0, 0, 0, 0]

// these are the ids of the arduinos that correspond to each controller
// This will let us correctly assing the right arduino to the right function
var pnpIds = {
  "USB\\VID_2341&PID_0042\\95635333331351416032" : SIMON_CENTER, 
  "USB\\VID_2341&PID_0042\\95635333331351801190" : SIMON_RED  


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

      readAllButtons();      
            
  
}

////////////////////////////////////
//
// Helper functions
//
//

// tells the ports what functions should handle each event
function setupHandlers(port) {
    console.log ("setting up ports");

    port.on('open', showPortOpen);
    port.on('data', receiveSerialData);
    port.on('close', showPortClose);
    port.on('error', showError);

}

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
  // console.log( String(data));
  //process.stdout.write(data);

  var btnId = -1;

   // gotta find a better way of doing this`
  if (data.toString().indexOf("SIMON_CENTER") > -1) {
    btnId = SIMON_CENTER;
  } else if (data.toString().indexOf("SIMON_RED") > -1) {
    btnId = SIMON_RED;
  } else if (data.toString().indexOf("SIMON_GREEN") > -1) {
    btnId = SIMON_GREEN;
  } else if (data.toString().indexOf("SIMON_BLUE") > -1) {
    btnId = SIMON_BLUE;
  } else if (data.toString().indexOf("SIMON_YELLOW") > -1) {
    btnId = SIMON_YELLOW;
  }

  switch ( btnId ) {
    case SIMON_CENTER:
      break;

    case SIMON_RED:
    case SIMON_GREEN:
    case SIMON_BLUE:
    case SIMON_YELLOW:
      if (data.toString().indexOf("BTTN:") > -1){
        var strIndex = data.toString().indexOf("BTTN:");
        buttonWeights[btnId] = parseInt(data.toString().substring(strIndex+5),10);
      }
      break;
  }


}

// reads one button
function readButton(btnId) {
  simonPorts[btnId].write ('READ_BUTTONS' + '\n');
}

// reads all readAllButtons
function readAllButtons() {

  simonPorts[SIMON_RED].write ('READ_BUTTONS' + '\n');
  simonPorts[SIMON_GREEN].write ('READ_BUTTONS' + '\n');
  simonPorts[SIMON_BLUE].write ('READ_BUTTONS' + '\n');
  simonPorts[SIMON_YELLOW].write ('READ_BUTTONS' + '\n');

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
const GS_INIT       = -1;
const GS_ATTRACT    = 1;
const GS_STARTED    = 2;

var gameState = GS_INIT;



// this will start the loop 5 seconds after we start up (to allow ports to be initialized), then
// start a 5 ms loop
setTimeout(function(){
  console.log("Starting Loop")
  setInterval (loop,5);
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
        gameState = GS_STARTED;
      }
      break;

    case GS_STARTED:
      simonPorts[SIMON_CENTER].write ('GS_COMPUTER' + '\n');
      break
  }


}
  