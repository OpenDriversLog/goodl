$(function() {
	pwRequired = true;
	initPasswordVerificators("resetPassword", "#reset");
});

function submitPwChangeForm() {

	if (!pwvalid || !pwrepeatvalid) {
		var errTxt = getErrorText();
		showWarning(errTxt);
		return;
	}
	$('#resetForm').submit();
}
