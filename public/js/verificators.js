var pwActive = false;
var pwRequired = false;
var pwvalid = false;
var pwrepeatvalid = false;
var mailRequired = false;
var mailvalid = false;
var prefix = "";
var hashprefix = "";
var mailRegexp = new RegExp('^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,4}$', 'i');
var requiredMissingCount = 0;
var requiredMissingFields = [];

function initPasswordVerificators(_prefix, _hashPrefix) {
    prefix = _prefix;
    hashprefix = _hashPrefix;
    if (!pwActive) {
        $(hashprefix + '_password').complexify({
            banMode: "mixed",
            strengthScaleFactor: 1
        }, checkPwValidity);
        pwActive = true;
        pwRequired = true;
        $(hashprefix + '_password2').bind('keyup focus input propertychange mouseup', checkPwRepeat);
        $(hashprefix + '_password').bind('keyup focus input propertychange mouseup', checkPwRepeat);
    }
}


function getErrorText() {
    var errTxt = "";
    if (!mailvalid && mailRequired) {
        errTxt += T(prefix + "_Errors_InvalidEmail") + "<br/>";
    }
    if (!pwvalid && pwRequired) {
        errTxt += T(prefix + "_Errors_InsecurePassword") + "<br/>";
    }
    if (!pwrepeatvalid && pwRequired) {
        errTxt += (T(prefix + "_Errors_passwordsNotMatch")) + "<br/>";
    }
    if (requiredMissingCount > 0) {
        $.each(requiredMissingFields, function(index, value) {
            errTxt += (T(prefix + "_Errors_" + value + "_missing")) + "<br/>";
        });
    }
    return errTxt;
}

function checkMail() {
    var mail = $(this).val();
    if (mail.match(mailRegexp)) {
      this.invalid=false;
        mailvalid = true;
    } else {
        this.invalid = true;
        mailvalid = false;
    }
    checkValid();
}

function checkRequired() {
    var val = $(this).val();

    if (val == "" || val == null) {
        setRequired(this);
    } else {
        $(this).toggleClass('valid', true);
        $(this).toggleClass('invalid', false);
        var arrPos = $.inArray($(this).attr('id'), requiredMissingFields);
        if (arrPos != -1) {
            // remove it
            requiredMissingFields.splice(arrPos, 1);
            requiredMissingCount--;
        }
    }
}

function setRequired(field) {
    $(field).toggleClass('valid', false);
    $(field).toggleClass('invalid', true);
    if ($.inArray($(field).attr('id'), requiredMissingFields) == -1) {
        requiredMissingFields.push($(field).attr('id'));
        requiredMissingCount++;
    }
}

function checkPwRepeat() {
    if (($(hashprefix + '_password2').val()) == ($(hashprefix + '_password').val())) {
        document.querySelector(hashprefix + '_password2').invalid = false;
        pwrepeatvalid = true;
    } else {
        document.querySelector(hashprefix + '_password2').invalid = true;
        pwrepeatvalid = false;
    }
    checkValid();
}

function checkPwValidity(valid, complexity) {
    var progressBar = $('#passwordbar');
    progressBar.toggleClass('progress-bar-success', valid && complexity >= 70);
    progressBar.toggleClass('progress-bar-danger', !valid);
    progressBar.toggleClass('progress-bar-warning', valid && complexity < 70);
    pwvalid = valid;
    if (valid) {
        document.querySelector(hashprefix + '_password').invalid = false;
    } else {
        document.querySelector(hashprefix + '_password').invalid = true;
    }
    if (complexity > 100) complexity = 100;
    progressBar.val(complexity);
    checkValid();
}

function checkValid() {
    var valid = (pwvalid || !pwRequired) && (pwrepeatvalid || !pwRequired) && (mailvalid || !mailRequired) && requiredMissingCount < 1;
    document.querySelector(hashprefix + "_submit").disabled = !valid;

    return valid;
}
