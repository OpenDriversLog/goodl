var curUsers = [];
var usersById = [];
var editRowsByUsrId = [];
var mailCBsWithPilot = [];
var mailCBsWithNewsletter = [];
var mailCBIdsWithPilot = [];
var mailCBIdsWithNewsletter = [];
var mailCBsWithNothing = [];

function GetUsers(finishFunc) {
	//  curContacts.sort(compareTrips);
	$('#contactsTablePlaceHolder').html('');
	showLoad();
	$.ajax({
		type: "POST",
		url: "./betaMan?action=read",
		async: true,
		dataType: 'json',
		success: function(result) {
			if (result.Error != false) {
				showError(result.ErrorMessage);
			} else {
				curUsers = result.BetaUsers;
				$.each(curUsers, function(key, val) {
					usersById[val.Id] = val;
				});

				refreshUsersTables();
			}
			hideLoad();
			finishFunc();
		},
		error: function(result) {
			showError(T("Error") + ": " + result.responseText);

			hideLoad();
			finishFunc();
		}
	});
}

function refreshUsersTables() {
	refreshMailUsersTable();
	refreshEditUsersTable();
}

function refreshMailUsersTable() {
	var tblBody = $('#sendMailTableBody');
	tblBody.html("");
	mailRowIdsWithPilot = [];
	mailRowIdsWithNewsletter = [];
	mailCBIdsWithPilot = [];
	mailCBIdsWithNewsletter = [];
	$.each(curUsers, function(k, usr) {
		usrTableIdx++;
		var curIdx = usrTableIdx;
		var row = $('<tr>');

		row.attr('id', 'sendMailRow' + curIdx);
		var cbId = "SendMailCB_" + usr.Id;
		var mailCb = $('<input type="checkbox" id="' + cbId + '" checked="checked" name="' + cbId + '"/>');
		var td = $('<td/>');
		var abc = $('<div class="input-field"/>');
		abc.append(mailCb);
		abc.append('<label for="' + cbId + '">Yes</label>');
		td.append(abc);

		row.append(td);
		row.append('<td>' + usr.Id + '</td>');
		row.append('<td>' + usr.Anrede + '</td>');
		row.append('<td>' + usr.Name + '</td>');
		row.append('<td>' + usr.Vorname + '</td>');
		row.append('<td>' + usr.Email + '</td>');
		row.append('<td>' + usr.Wants2BePilot + '</td>');
		row.append('<td>' + usr.WantsNewsletter + '</td>');

		if (usr.Wants2BePilot == 1) {
			mailCBsWithPilot.push(mailCb);
			mailCBIdsWithPilot.push(cbId);
		}
		if (usr.WantsNewsletter == 1) {
			mailCBsWithNewsletter.push(mailCb);
			mailCBIdsWithNewsletter.push(cbId);

		}
		if (!usr.Wants2BePilot && !usr.WantsNewsletter) {
			mailCBsWithNothing.push(mailCb);
		}
		tblBody.append(row);
		row.show();
	});
}

function GetBetaUserFromRow(rowIdx, onlyIfEnabled) {
	var row = $("#editUsersTable").find('#betaUserRow' + rowIdx);
	var contact = dynamicFillObject(row, '.userItem', onlyIfEnabled);

	return contact;
}
var usrTableIdx = 0;

function refreshEditUsersTable() {
	var tblBody = $('#editUsersBody');

	var rowDummy = tblBody.find('#editUserDummyRow');
	tblBody.html("");
	tblBody.append(rowDummy);
	$.each(curUsers, function(k, usr) {
		var id = usr.Id;
		usrTableIdx++;
		var curIdx = usrTableIdx;
		var row = rowDummy.clone();

		row.attr('id', 'betaUserRow' + curIdx);
		dynamicFillRow(row, usr, '.userItem');


		row.find('[name="Id"]').val(id);
		row.find('select').attr('id', 'selusr' + curIdx);

		tblBody.append(row);

		row.find('[name=SaveUser]').bind('click', function() {
			SaveUser(curIdx);
		});
		row.find('[name=DeleteUser]').bind('click', function() {
			DeleteUser(curIdx);
		});

		row.find('select').material_select();

		row.find('.unmodified').bind('keyup focus input propertychange mouseup', function() {

			t = $(this);
			t.toggleClass('unmodified', false);
			t.removeAttr('readonly');
		});

		editRowsByUsrId[row.find('[name="Id"]').val()] = row;
		row.show();
	});
}

function SaveUser(_rowIdx) {

	var rowIdx = _rowIdx;


	usr = GetBetaUserFromRow(rowIdx);
	action = "update";
	if (usr.Id == 0) {
		showError("No usrId found - cancel!");
		return;
	}
	executeUserCommand(action, usr, function(result, err, action) {
		if (err == null) {
			var row = $("#editUsersBody").find('#betaUserRow' + rowIdx);
			row.fadeIn(100).fadeOut(100).fadeIn(100);
		} else {
			showError(err);
		}
	});


}

function DeleteUser(rowIdx) {
	usr = GetBetaUserFromRow(rowIdx);
	action = "delete";
	if (usr.Id == 0) {
		showError("No usrId found - cancel!");
		return;
	}
	executeUserCommand(action, usr, function(result, err, action) {
		if (err == null) {
			var row = $("#editUsersBody").find('#betaUserRow' + rowIdx);
			row.hide("slow");
		} else {
			showError(err);
		}
	});
}

function executeUserCommand(action, usr, finishFunc) {
	data = {};
	data["action"] = action;
	data["betaUser"] = JSON.stringify(usr);
	console.log("ExecuteUserCommand");
	console.log(data);
	$.ajax({
		type: "POST",
		url: "./betaMan",
		async: true,
		data: data,
		dataType: 'json',
		success: function(result) {
			if (!result.Success) {
				finishFunc(result, result.ErrorMessage, action);
			} else {
				finishFunc(result, null, action);
			}
			hideLoad();
		},
		error: function(result) {
			showError(T("Error") + ": " + result.responseText);
			hideLoad();
		}
	});
}
$(function() {
	GetUsers(function() {
		refreshUsersTables();
	});
	$('#wants2BeCb').change(function() {
		updateAutoCheck();
	});
	$('#wantsNewsLetterCB').change(function() {
		updateAutoCheck();
	});
});

function updateAutoCheck() {
	var pilot = $('#wants2BeCb').is(':checked');
	var news = $('#wantsNewsLetterCB').is(':checked');
	$.each(mailCBsWithPilot, function(k, cb) {
		cb.prop('checked', pilot || (news && (mailCBIdsWithNewsletter.indexOf(cb.attr('id')) != -1)));

	});
	$.each(mailCBsWithNewsletter, function(k, cb) {
		cb.prop('checked', news || (pilot && (mailCBIdsWithPilot.indexOf(cb.attr('id')) != -1)));
	});
	$.each(mailCBsWithNothing, function(k, cb) {
		cb.prop('checked', false);
	});
}