async function Logout() {
    let cfg = { method: 'post' }
    const res = await fetch('/client/logout', cfg).catch(err => {
        console.log(err)
        return
    })
    if (res.status == 200) {
        window.location.href = "/client/login";
    } else {
        console.log(res.statusText)
    }
}