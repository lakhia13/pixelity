//
// Filename: navbar.tsx
// Description: The website's nav bar
// Created: 3/29/25
//

"use client";

import { Button, Card } from "pixel-retroui";
import { JSX } from "react";
import Image from "next/image";
import { MdAccountCircle } from "react-icons/md";

export default function NavBar(): JSX.Element {
	return (
		<div className="flex p-4 gap-x-4 bg-red-400 rounded-b-4">
			<Card>
				<Image src="/logo.png" alt="Pixelity Logo" width={75} height={75}/>
			</Card>
			
			<div className="rounded-lg bg-white border-0 border-b-2 border-solid border-b-red-600 text-black w-full flex justify-center items-center">
				<p className="text-7xl text-center">Pixelity</p>
			</div>

			<Button onClick={void(0)}> <MdAccountCircle size={65}/> </Button>
		</div>
	);
} 