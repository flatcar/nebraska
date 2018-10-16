/*
Simple server side sessions for goji

	goji.Use(Sessions.Middleware())

In-memory session store:

	var secret               = "thisismysecret"
	var inMemorySessionStore = sessions.MemoryStore{}
	var Sessions             = sessions.NewSessionOptions(secret, &inMemorySessionStore)

Using Redis (using fzzy/radix):

	var redisSessionStore = sessions.NewRedisStore("tcp", "localhost:6379")
	var Sessions          = sessions.NewSessionOptions(secret, redisSessionStore)

Use middleware:

	goji.Use(Sessions.Middleware())

Accessing session variable:

	func handler(c web.C, w http.ResponseWriter, r *http.Request) {
		sessionObj := Sessions.GetSessionObject(&c)

        // Regnerate session..
        Sessions.RegenerateSession(&c)

        // Delete session
        Sessions.DeleteSession(&c)
	}

See examples folder for full example.
*/
package sessions
