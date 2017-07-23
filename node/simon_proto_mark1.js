var simonPorts = [];

var SerialPort = require('serialport');
// callback approach 
SerialPort.list(function (err, ports) {
  ports.forEach(function(port) {
    console.log(port.comName);
    console.log(port.pnpId);
    console.log(port.manufacturer);
    var newPort = new SerialPort (port.comName,{
      baudRate: 115200,
      autoOpen: false
    });
    simonPorts.push (newPort);
    setupHandlers(newPort);
    openPort (newPort);
  });
});


function setupHandlers(port) {
    console.log ("setting up ports");

    port.on('open', showPortOpen);
    port.on('data', sendSerialData);
    port.on('close', showPortClose);
    port.on('error', showError);

}

function openPort (port) {
  console.log(port)
  port.open(function(error) {
      if (error) {
        console.log('failed to open port: ' + error);
      }
  });
}




function showPortOpen(path, options) {
  console.log('port open. Data rate: ' + path);
}

function stripAlphaChars(source) {
  var out = source.replace(/[^0-9]/g, '');

  return out;
}

function sendSerialData(data) {
  // console.log( String(data));
  //process.stdout.write(data);

  if (data.toString().indexOf("SIMON_RED") > -1) {
    if (data.toString().indexOf("BTTN:") > -1){
        var strIndex = data.toString().indexOf("BTTN:");
        simonRedBttnWeight = parseInt(data.toString().substring(strIndex+5),10);


    }

  }

}
 
function showPortClose() {
   console.log('port closed.');
}
 
function showError(error) {
   console.log('Serial port error: ' + error);
}


/*
 * Main Simon Game State Loop
 */

const SIMON_CENTER = 0;
const SIMON_RED    = 1;
const SIMON_GREEN  = 2;
const SIMON_BLUE   = 3;
const SIMON_YELLOW = 4;

var ledNum = 0;
var maxLEDS = 144;
var arduinoString = "";
var gameState = -1;

var simonRedBttnWeight = 0;

setTimeout(function(){
  setInterval (loop,5);
}, 5000);


function loop () {

  if (gameState == -1) {
    // setup parts
    simonPorts[SIMON_CENTER].write ('SIMON_CENTER' + '\n');
    simonPorts[SIMON_RED].write ('SIMON_CENTER' + '\n');
    simonPorts[SIMON_CENTER].write ('GS_ATTRACT' + '\n');
    simonPorts[SIMON_RED].write ('GS_ATTRACT' + '\n');
    gameState = 1;
  } else if (gameState == 1) {
    
    //simonPorts[SIMON_RED].write ('READ_BUTTONS' + '\n');
    if (simonRedBttnWeight > 0) {
      console.log ("Button Pushed :: Start Game");
      gameState = 2;
    }
  } else if (gameState == 2) {
    simonPorts[SIMON_CENTER].write ('GS_COMPUTER' + '\n');
  }


}
