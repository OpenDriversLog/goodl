var lastErrors = [];
var lastStatusses = [];
var lastWarnings = [];
var navbaroffset = 72;
var translations = null;
var notLoggedInPages=[
	"resetPassword",
	"newPassword",
	"register",
	"login",
	"impressum",
	"alpha",
	"beta-dev",
	"test"
];
if (typeof String.prototype.startsWith != 'function') {
	String.prototype.startsWith = function (str){
		return this.slice(0, str.length) == str;
	};
}

if (typeof String.prototype.endsWith != 'function') {
	String.prototype.endsWith = function (str){
		return this.slice(-str.length) == str;
	};
}

function T(key) {
	if (main) {
		return main.t[key];
	}
	return key;
}

function appReady(app, func) {

	if(!app.odlLoaded) {
		if(!app.odlInjected) {
			app.odlInjected = true;
			app.addEventListener('dom-change', function() {
				app.odlLoaded = true;
			});
		}
		app.addEventListener('dom-change', function() {
			func.call(app);
		});
	} else {
		func();
	}

}
var userData;
function initUser(ud) {
	userData = ud;
	console.log("Paul - hier hast Du deinen User" + userData["LoginName"] + ". Mach was draus (externes Skript bla blubb, im View ist auch schon Zeug ausgeblendet ohne User).");

}

function showResultErrors(result,t) {

	console.warn("Showing result errors", result);
	var msg = "";
	if (result.Errors == null || result.Errors.length == 0 || result.Errors.length == undefined) {
		msg = T("Error") + ": " + getTransMsg(result.ErrorMessage,t);
	} else {
		$.each(result.Errors, function(index, value) {
			msg += getTransMsg(value,t);
			msg += "<br/>";
			$('#' + index).toggleClass('invalid', true);
			$('#' + index).click(function() {
				$(this).toggleClass('invalid', false);
			});
		});
	}
	showWarning(msg);

}

function showLoad(tag) {
	if(tag==undefined) tag="default";
	if(document.getElementById('loadFinished')==null) {
		appReady(app,function(){showLoad(tag);});
	} else {
		load(tag); // defined in odl-layout.html
	}
}

function hideLoad(tag) {
	if(tag==undefined) tag="default";
	if(document.getElementById('loadFinished')==null) {
		appReady(app,function(){showLoad(tag);});
	} else {
		unload(tag); // defined in odl-layout.html
	}
}

function showError(message,t) {
	hideError();
	message = getTransMsg(message,t);

	if(T("UseReportForm")==undefined) {
		window.setTimeout(function(){
			showError(message,t);
		},333);
		return;
	}
	$("#eb").html(message + "<br/><b>" + T("UseReportForm") + "</b>");
	console.warn("show error!");
	appReady(app,function() {
		window.setTimeout(function(){
			document.getElementById("error").open();
		},200);
	});

	lastErrors.push($('#eb').text());
}

function getTransMsg(message,trans,dontTryMain) {
	if(!message) return "";
	var t;
	if(trans===undefined) {
		t = main.t;
		dontTryMain=true;
	} else {
		t = trans;
	}
	if(!t) {
		console.warn("No t found");
		return message;
	}
	lmessage = message.toLowerCase();
	smessage = t[lmessage];
	if (smessage && smessage != "") {
		return smessage;
	}
	smessage = t[message];
	if (smessage && smessage != "") {
		return smessage;
	}
	if(!dontTryMain) {
		message = getTransMsg(message,main.t,true);
	}
	return message;
}

function showStatus(message,t) {

	hideStatus();
	message = getTransMsg(message,t);
	$("#sb").html(message);

	document.getElementById("status").open();
	lastStatusses.push($('#sb').text());

}

function showWarning(message,t) {
	hideWarning();
	message = getTransMsg(message,t);
	$("#wb").html(message);
	document.getElementById("warning").open();
	lastWarnings.push($("#wb").text());
}

function isEmpty(str) {
	return (!str || 0 === str.length || !str.trim());
}

function initDefault(app) {
	var er = $('#eb').text();
	if (!isEmpty(er)) {
		lastErrors.push(er);
	}
	var w = $('#wb').text();
	if (!isEmpty(w)) {
		lastWarnings.push(w);
	}
	var s = $('#sb').text();
	if (!isEmpty(s)) {
		lastStatusses.push(s);
	}
}

function setRoute(route) {
	main.set("route.path",main.route.path.substring(0,main.route.path.indexOf("/odl/"))+"/odl"+route);
	main.set("tail.path","");
}

function submitErrorReport() {

	showLoad();
	var data = {};
	var txt = $('#reportErrorTextArea').val();
	if ($("#sendMetaData")[0].checked) {

		txt += "\r\n--------------------------------------------\r\n" + T("metaData") +
			"\r\n Zeitstempel (Seitenaufruf): " + timeLoad + "\r\n Seitenadresse : " + window.location.href + "\r\n\r\nFehler : ";
		$.each(lastErrors, function(k, v) {
			txt += "\r\n" + v;
		});
		txt += "\r\n\r\nWarnungen : ";
		$.each(lastWarnings, function(k, v) {
			txt += "\r\n" + v;
		});
		txt += "\r\n\r\nStatus : ";
		$.each(lastStatusses, function(k, v) {
			txt += "\r\n" + v;
		});
		txt+="\r\n\r\n -------- Browser-Info ------------"
		txt += "\r\nUserAgent : " + navigator.userAgent;
		txt += "\r\nVendor : " + navigator.vendor;
		txt += "\r\nappName : " + navigator.appName;
		txt += "\r\nappCodeName : " + navigator.appCodeName;
		txt += "\r\nappVersion : " + navigator.appVersion;
		txt += "\r\nplatform : " + navigator.platform;
	};

	data["text"] = $('<div/>').text(txt).html().replace(/\n/g, "<br/>"); // escaped
	$.ajax({
		type: "POST",
		url: "./sendMail",
		data: data,
		async: true,
		success: function(result) {
			// if (initialRefreshCount == refreshCount) {
			if (result == "Success") {
				document.getElementById('thanksToast').show();

				showStatus(T("Report_successful"));
			} else {
				showError(T("Error") + ": " + result);

			}
			hideLoad();
			// }
		},
		error: function(result, status) {
			if (result.responseText != undefined && result.responseText.indexOf('<!DOCTYPE html>')==-1) {
				showError(T("Error") + ": " + result.responseText);
			} else {
				showError(T("Error") + ": " + result.status + " - " + status);
			}
			hideLoad();
		}
	});
}

function callWithMessage(url, statusMessage, warningMessage, errorMessage) {
	var form = $('<form action="' + url + '" method="post" style="display:none">' +
		'<input type="text" name="statusMessage" value="' + (statusMessage == undefined ? "" : statusMessage) + '" />' +
		'<input type="text" name="warningMessage" value="' + (warningMessage == undefined ? "" : warningMessage) + '" />' +
		'<input type="text" name="errorMessage" value="' + (errorMessage == undefined ? "" : errorMessage) + '" />' +
		'</form>');
	$('body').append(form);
	form.submit();
}

function hideMessages() {
	hideError();
	hideWarning();
	hideStatus();
}

function hideWarning() {
	document.getElementById("warning").close();
}

function hideStatus() {
	document.getElementById("status").close();
}

function hideError() {
	document.getElementById("error").close();
}
function reportError() {
	document.getElementById('reportErrorDialog').open();
}
/**
 * Gets the index of the item with the given id
 * @param {number} id The id to search
 * @param {array} arr The array to scan
 * @returns {number}
 */
function getIdxFromId(id, arr) {
	for (var i = 0; i < arr.length; i++) {
		if (arr[i].Id == id) {
			return i;
		}
	}
	return -1;
}

function handleAjaxError(e,el,t) {
	var err = e.detail.error;
	var uri = e.detail.request.url;
	if (err && err.message=="Request aborted.") {
		console.log("Request " + uri + " got canceled.");
	} else {
		console.error("Error occured on ajax-request " + uri +" : ", err);
		lastErrors.push("Error on ajax-request for " + uri);
		showError("Es ist ein Fehler beim Anfordern von Daten aufgetreten - bitte aktualisieren Sie die Seite und versuchen es erneut. Sollte das Problem weiterhin bestehen, kontaktieren Sie uns bitte.")
	}
}

function getReplacedAjaxUri(fetcherUri) {
	if(main.ajaxUri) {
		var repUri = document.location.origin;
		if (main.subdirToReplace) {
			var sdRep = main.subdirToReplace;
			if(main.subdirToReplace.indexOf("***")===0) {
				repUri = fetcherUri.substring(0,fetcherUri.indexOf(sdRep.substring(3))-3+sdRep.length);
			} else {
				repUri += "/" + sdRep;
			}

		}
		fetcherUri = fetcherUri.replace(repUri, main.ajaxUri)
	}
	return fetcherUri;
}