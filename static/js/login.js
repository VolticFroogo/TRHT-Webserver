$(document).ready(function(){
    Waves.displayEffect();

    $("#login-button").click(function(){
        Materialize.toast('Sending login request!', 4000);

        $.ajax({
            url: "/loginajax",
            type: "post",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                Email: $("#email").val(),
                Password: $("#password").val(),
                Captcha: grecaptcha.getResponse()
            }),
            dataType: "json",
            success: function(r) {
                if (r.success) {
                    window.location.replace("/admin");
                } else {
                    Materialize.Toast.removeAll(); // Clear all other toasts.
                    Materialize.toast('Invalid login credentials.', 4000);
                }
            }
        });

        grecaptcha.reset(); // Reset the recaptcha
    });
});