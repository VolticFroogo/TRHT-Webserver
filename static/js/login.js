$(document).ready(function(){
    Waves.displayEffect();

    $("#login-button").click(function(){
        $.ajax({
            url: "/loginajax",
            type: "post",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                Email: $("#email").val(),
                Password: $("#password").val()
            }),
            dataType: "json",
            success: function(r) {
                if(r.success) {
                    window.location.replace("/admin");
                }
            }
        });
    });
});