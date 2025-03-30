//
// Filename: page.tsx
// Route: /signin
// Created: 3/29/25
//

"use client";

import { Button, Input } from "pixel-retroui";
import Form from "next/form"
import { FormEvent, JSX, useState } from "react";
import { signIn } from "next-auth/react";
import { HiEye, HiEyeOff } from "react-icons/hi";

interface SignInProps {
	searchParams?: Promise<{ [key: string]: string | string[] | undefined }>
}

export default function SignIn({ searchParams }: SignInProps): JSX.Element {
	const [viewPassword, setViewPassword] = useState(false);
	const [authError, setAuthError] = useState<string | null>(null);
	const [username, setUsername] = useState("");
	const [password, setPassword] = useState("");
	
	const viewIcon = () => viewPassword ? <HiEyeOff/> : <HiEye/>;
	const inputType = () => viewPassword ? "text" : "password";

	const submit = async (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault(); // Prevent default form submission
		
		try {
			const params = await searchParams;

			// Use a default redirect if none provided
			const callbackURL = params && params.url? params.url as string : "/view";

			const response = await signIn("credentials", {
				username,
				password,
				redirect: true,
				callbackUrl: callbackURL	// Redirect to where the user was attempting to visit
			});
			
			if (!response || response.error == "CredentialsSignin") {
				setAuthError("Invalid username or password");
			}
		}
		catch (error) {
			console.error("Sign-in error:", error);
			setAuthError("An error occurred while signing in. Please try again later.");
		}
	};
	
	return (
		<div
			className={`
				absolute inset-0 -z-10 h-full flex items-center justify-center
				font-[family-name:var(--font-geist-sans)]

			`}
		>
			<main className="bg-red-400 rounded-[16px] flex flex-col w-[500px] gap-y-8 p-10">
				<h1 className="text-2xl text-white font-bold">Sign In</h1>
				
				<Form
					id="sign-in"
					autoComplete="on"
					action=""
					onSubmit={submit}
				>
					<div className="flex flex-col gap-y-4 w-full">
						<Input
							className="text-black bg-white p-4 rounded-lg"
							placeholder="Enter your username"
							type="text"
							onChange={ (event) => setUsername(event.target.value) }
						/>
						
						<div className="relative flex">
							<Input
								className="text-black bg-white p-4 rounded-lg w-full"
								placeholder="Enter your password"
								type={inputType()}
								onChange={ (event) => setPassword(event.target.value) }
							/>

							<Button className="absolute right-0 top-1 text-gray-400 bg-white p-4 rounded-lg" onClick={ () => setViewPassword(prev => !prev) }>
								{ viewIcon() }
							</Button>
						</div>
						
						{authError && <p className="text-danger text-sm">{authError}</p>}
					</div>
				</Form>
				
				<div className="flex gap-x-2 justify-end">
					<Button
						type="submit"
						form="sign-in"
						color="primary"
					>
						Sign In
					</Button>
				</div>
			</main>
		</div>
	);
}