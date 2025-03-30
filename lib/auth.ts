//
// Filename: auth.ts
// Description: NextAuth options
// Created: 3/29/25
//

import CredentialsProvider from "next-auth/providers/credentials";
import { AuthOptions } from "next-auth";

export const authOptions = {
	providers: [
		CredentialsProvider({
			name: "Credentials",
			credentials: {
				username: { label: "Username", type: "text" },
				password: { label: "Password", type: "password" },
			},

			async authorize(credentials) {
				if (!credentials) {
					return null;
				}

				console.log("Credentials: ", credentials);

				return {
					id: "",
					username: credentials.username,
					password: credentials.password,
				};
			},
		}),
	],

	pages: {
		signIn: "/signin",
		newUser: "/signup",
	},
	
	// Write custom callbacks to deal with the custom session and user data
	callbacks: {
		async redirect({ url, baseUrl }: { url: string, baseUrl: string }): Promise<string> {
			if (url.startsWith("/")) {						// Allows relative callback URLs
				return `${baseUrl}${url}`;
			}
			else if (new URL(url).origin === baseUrl) {	// Allows callback URLs on the same origin
				return url;
			}

			return baseUrl;
		},

		async session({ session, token }: { session: any, token: any }): Promise<any> {
			if (token) {
				session.user.id 		= token.id;
				session.user.username 	= token.username;
				session.user.password 	= token.password;
			}
			
			return session;
		},

		async jwt({ token, user }: { token: any, user?: any }): Promise<any> {
			if (user) {
				token.id 		= user.id;
				token.username 	= user.username;
			}
			
			return token;
		},
	},

	session: {
		strategy: 	"jwt",
		maxAge: 	5184000, 	// Make users reauthenticate every 60 days
		updateAge: 	43200, 		// Update session every 12 hours
	},

	jwt: {
		maxAge: 	5184000, 	// Refresh the JWT session every 60 days
	},

	secret: process.env.NEXTAUTH_SECRET
} satisfies AuthOptions;