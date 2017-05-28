"use strict";
(function() {
	var socket = new  WebSocket("ws://"+location.host+"/ws");

	function strlen(str){
   var length = str.length, count = 0, i = 0, ch = 0;
   for(i; i < length; i++){
     ch = str.charCodeAt(i);
     if(ch <= 127){
        count++;
     }else if(ch <= 2047){
        count += 2;
     }else if(ch <= 65535){
        count += 3;
     }else if(ch <= 2097151){
        count += 4;
     }else if(ch <= 67108863){
        count += 5;
     }else{
        count += 6;
     }    
  }
  return count;
};

	function submitHandle() {

		var msg = $('#message-text').val();

		$('#message-text').val('');;

		var len = strlen(msg);

		var slen = len.toString();

		while (slen.length < 8) slen += ' ';

		socket.send(slen + msg);
	}

	function establishSocket() {

		socket.onopen = function() {
			console.log("Connected");
		}

		socket.onclose = function(event) {
			if (event.wasClean) {
	    		console.log('Connection closed');
		  	} else {
		    	console.log('Error: Connection reset');
		  	}
		  		console.log('Code: ' + event.code + ' reason: ' + event.reason);
		  		var n = $('#messages').height()	
		  		$('#messages').append('<li class="list-group-item list-group-item-danger short-text">Connection lost at '+new Date()+'</li>').scrollTop(n + 9999);

		}

		socket.onmessage = function (msg) {
			var jmsg = JSON.parse(msg.data, function(key, value) {
				if (key == 'messagetime') {
					var ms = Date.parse(value);
					return new Date(ms); 
				}
				return value;
			})
			switch(jmsg.messagetype) {
				case 'Admin':
					var n = $('#messages').height()		
					$('#messages').append('<li class="list-group-item list-group-item-info short-text">'+jmsg.username+' : '+jmsg.messagetext+' at '+jmsg.messagetime+'</li>').scrollTop(n + 9999);
					break;
				case 'User':
					var n = $('#messages').height()	
					$('#messages').append('<div class="panel panel-primary"><div class="panel-heading short-text">'+jmsg.username+' at '+jmsg.messagetime+'</div><div class="panel-body short-text">'+jmsg.messagetext+'</div></div>').scrollTop(n + 9999);
					break;
				default:
					break;
			}
		};

		socket.onerror = function(error) {
		 	console.log("Ошибка " + error.message);
		};
	}

	establishSocket();

	window.SMB = submitHandle
  
})();