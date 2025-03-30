//
// Filename: page.tsx
// Route: /
// Created: 3/29/25
//

"use client";

import { Bubble, Button, Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "pixel-retroui";
import Image from "next/image";
import { SessionProvider } from "next-auth/react";
import { useState } from "react";

import AuthPopup from "@/components/auth-popup";

import '@/lib/pixel-retroui-setup.js';

export default function Home() {
	const [isOpen, setIsOpen] = useState(false);
	
	const togglePopup = () => setIsOpen((prev) => !prev);
	
	return (
		<div className="my-auto">
			<main>
				<div className="flex justify-center items-center w-full">
					<div className="justify-start items-start my-[300px] flex flex-col w-1/3">
						<Bubble
							className="flex justify-center items-center"
							direction="left"
							onClick={ () => {
								console.log("Clicked");
							}}
						>
							Your Images, Under Your Control
						</Bubble>
						<Image src="/mascot_no_bg.png" alt="Pixelity Logo" width={300} height={300}/>
					</div>
					
					<div className="justify-end items-end my-[300px] w-1/4 flex flex-col">
						<Accordion>
							<AccordionItem value="item-1">
								<AccordionTrigger>Why use Pixelity</AccordionTrigger>
								
								<AccordionContent>
									If you have an old android phone lying around, with no use than gathering dust. And if you 
									don&apos;t want to hand over your photos to Big Data, then you have come to the right place.
								</AccordionContent>
							</AccordionItem>

							<AccordionItem value="item-2">
								<AccordionTrigger>How does it Work</AccordionTrigger>

								<AccordionContent>
									Our android app interfaces with our web login. You can use your old phone as your personal,
									private cloud server. Your photos will be stores on your device, accessible to you only,
									from anywhere you want! 
								</AccordionContent>
							</AccordionItem>
						</Accordion>
					</div>

					<SessionProvider>
						<AuthPopup isOpen={isOpen} onClose={togglePopup}/>
					</SessionProvider>
					
					<Button onClick={togglePopup}>Open Popup</Button>
				</div>
			</main>
		</div>
	);
}
