//
// Filename: auth-popup.tsx
// Description: The authentication popup
// Created: 3/30/25
//

"use client";

import { Button, Input, Popup } from "pixel-retroui";
import { BuiltInProviderType } from "next-auth/providers/index";
import { ClientSafeProvider, getProviders, LiteralUnion, signIn, useSession } from "next-auth/react"
import Form from "next/form"
import { HiEye, HiEyeOff } from "react-icons/hi";
import { JSX, FormEvent, useEffect, useRef, useState } from "react";

interface AuthPopupProps {
	isOpen: 	boolean;
	onClose: () => void;
}

type Providers = Record<LiteralUnion<BuiltInProviderType, string>, ClientSafeProvider>

export default function AuthPopup({ isOpen, onClose }: AuthPopupProps): JSX.Element {
	const [viewPassword, setViewPassword] = useState(false);
	const [authError, setAuthError] = useState<string | null>(null);
	const [username, setUsername] = useState("");
	const [password, setPassword] = useState("");
	const [providers, setProviders] = useState<Providers | null>(null);
	
	const { data: session } = useSession();
	
	const sessionRef = useRef(session);

	useEffect(() => {
		(async () => {
			setProviders(await getProviders())
		})();
	}, []);

	useEffect(() => {
		console.log(providers);
	}, [providers]);

	const viewIcon = () => viewPassword? <HiEyeOff/> : <HiEye/>;
	const inputType = () => viewPassword? "text" : "password";

	const submit = async (event: FormEvent<HTMLFormElement>) => {
		if (!providers) {
			return;
		}
		
		event.preventDefault(); // Prevent default form submission

		try {
			const response = await signIn("credentials", {
				username,
				password,
				redirect: true,
				callbackUrl: "/api/auth/callback/credentials"
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
		<Popup
			isOpen={isOpen}
			onClose={onClose}
		>

			<Form
				className="w-[500px]"
				id="sign-in"
				autoComplete="on"
				action=""
				onSubmit={submit}
			>
				<div className="flex flex-col gap-y-4">
					<div className="flex flex-col gap-y-4 w-full"> {
						providers && Object.values(providers).map((provider) => (
							provider.name !== "Credentials" && (
								<Button
									key={provider.name}
									onClick={() => signIn(provider.id, { callbackUrl: "/" })}
								>
									{`Sign in with ${provider.name}`}
								</Button>	
							)
						))
					} </div>
					
					<div className="flex flex-col gap-y-4 w-full">
						<Input
							className="text-black bg-white rounded-lg"
							placeholder="Enter your username"
							type="text"
							onChange={ (event) => setUsername(event.target.value) }
						/>
						
						<div className="relative flex">
							<Input
								className="text-black bg-white rounded-lg w-full"
								placeholder="Enter your password"
								type={inputType()}
								onChange={ (event) => setPassword(event.target.value) }
							/>

							<div className="hover:bg-red-500 cursor-pointer" onClick={ () => setViewPassword(prev => !prev) }>
								<span className="absolute right-5 top-5 text-gray-400 bg-white rounded-lg">
									{ viewIcon() }
								</span>
							</div>
						</div>
						
						{authError && <p className="text-danger text-sm">{authError}</p>}
					</div>
				</div>
			</Form>
			
			<div className="flex relative pt-2 justify-end">
				<Button type="submit" form="sign-in" color="primary">
					Sign In
				</Button>
			</div>
		</Popup>
	);
}