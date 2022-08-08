const button = document.getElementById("login");

button.addEventListener("click", async function() {
    const username = document.getElementById('usernameInput').value
    const pass = document.getElementById('passwordInput').value

    let cfg = {
        method: 'post',
        body: JSON.stringify({
            username: username,
            password: pass
        })
    }
    const res = await fetch('/auth/login', cfg).catch(err => {
        console.log(err)
        return
    })
    if (res.status != 200) {
        showError(res.statusText)
    } else {
        window.location.href = "/client";
    }
});

function showError(error) {
    const errdiv = document.getElementsByClassName('error')[0]
    errdiv.innerHTML = "<div class=\"alert alert-dismissible alert-danger\">\n" +
        "  <h4 class=\"alert-heading\">Failed to log in.</h4>\n" +
        "  <p class=\"mb-0\">Logging in produced an error: " + error + "</p>\n" +
        "</div>"
}