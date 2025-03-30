//
// Filename: page.tsx
// Route: /
// Created: 3/29/25
//

"use client";

import { Button, Bubble, Card } from "pixel-retroui";
import Image from "next/image";
import { MdAccountCircle } from "react-icons/md";

import '@/lib/pixel-retroui-setup.js';

export default function Home() {
	return (
		<div>
			<main>
				<div className="flex p-4 gap-x-4 bg-red-400 rounded-b-4">
					<Card>
						<Image src="/logo.png" alt="Pixelity Logo" width={75} height={75}/>
					</Card>
					
					<div className="rounded-lg bg-white border-0 border-b-2 border-solid border-b-red-600 text-black w-full flex justify-center items-center">
						<p className="text-7xl text-center">Pixelity</p>
					</div>

					<Button onClick={void(0)}> <MdAccountCircle size={65}/> </Button>
				</div>

				<div className="justify-center items-center my-[300px] flex flex-col">
					<Bubble
						className="flex justify-center items-center w-1/4"
						direction="left"
						onClick={ () => {
							console.log("Clicked");
						}}
					>
						Upload your image
					</Bubble>
				</div>
			</main>
		</div>
	);
}
