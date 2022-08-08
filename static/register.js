const button = document.getElementById("register");

button.addEventListener("click", async function() {
    const user = document.getElementById('usernameInput').value
    const email = document.getElementById('emailInput').value
    const pass = document.getElementById('passwordInput').value

    let cfg = {
        method: 'post',
        body: JSON.stringify({
            username: user,
            email: email,
            password: pass
        })
    }
    const res = await fetch('/register', cfg).catch(err => {
        console.log(err)
        return
    })
    if (res.status != 200) {
        showError(res.statusText + "\n" + await res.text())
    } else {
        window.location.href = "/login";
    }
});

function showError(error) {
    const errdiv = document.getElementsByClassName('error')[0]
    errdiv.innerHTML = "<div class=\"alert alert-dismissible alert-danger\">\n" +
        "  <h4 class=\"alert-heading\">Failed to register.</h4>\n" +
        "  <p class=\"mb-0\">Logging in produced an error: " + error + "</p>\n" +
        "</div>"
}