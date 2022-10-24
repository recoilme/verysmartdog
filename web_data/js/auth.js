function onTelegramAuth(user) {
    console.log('New user:', user)
    load(user)
}

function bake_cookie(name, value) {
    var cookie = [name, '=', JSON.stringify(value), '; domain=.', window.location.host.toString(), '; path=/;'].join('');
    document.cookie = cookie;
}

async function load(tgUser) {
    const client = new PocketBase('http://127.0.0.1:80');

    errAuth = 0;
    try {
        const user = await client.users.create({
            'email': tgUser.id +'@t.me',
            'password': '123456'+tgUser.username,
            'passwordConfirm': '123456'+tgUser.username,
        });
    } catch (e) {
        if (e.data.data.email.code=="validation_user_email_exists") {
            errAuth = 42;
        } else {
            errAuth = 43;
        }
    }
    if (errAuth == 42 || errAuth == 0) {
        // user authentication via email/pass
        const userAuthData = await client.users.authViaEmail(tgUser.id +'@t.me', '123456'+tgUser.username);
        console.log('userAuthData:', userAuthData)

        // set user profile data
        const updatedProfile = await client.records.update('profiles', userAuthData.user.profile.id, {
        'name': tgUser.username,
        'photo_url': tgUser.photo_url,
        });
        console.log('updatedProfile:', updatedProfile)

        bake_cookie("t",userAuthData.token)

        document.location.href="/"
    }
    
}