//
// Filename: page.tsx
// Route: /
// Created: 3/29/25
//

"use client";

import { Button } from 'pixel-retroui';
import Link from "next/link";
import { useState, JSX } from 'react';

import '@/lib/pixel-retroui-setup.js';

export default function View() {
    const [selectedButton, setSelectedButton] = useState('List');
    const [view, setView] = useState('List');

	const onClick = (buttonName: string) => {
		setSelectedButton(buttonName);
        setView(buttonName);
		console.log(`Clicked ${buttonName}`);
	};

    const ViewMode = ():JSX.Element | null =>{
        if (view === 'Albums') {
            // Render the albums view
            return (
                <div className="pl-8 py-3">
                    <h1 className="text-2xl">Albums View</h1>
                    {/* Add your albums view content here */}
                </div>
            );
        } else if (view === 'List') {
            // Render the all view
            return (
                <div className="pl-8 py-3">
                    <h1 className="text-2xl">List View</h1>
                    {/* Add your all view content here */}
                </div>
            );
        } else if (view === 'Grid') {
            // Render the all view
            return (
                <div className="pl-8 py-3">
                    <h1 className="text-2xl">Grid View</h1>
                    {/* Add your all view content here */}
                </div>
            );
        } else {
            // Render a default view or return null
            return null;
        }
    };
	return (
		<div>
			<main className='relative flex'>
                    <Link href="/upload">
                        <Button
                            className='bg-gray-300'
                            style={{
                                backgroundColor: '#e5e7eb',
                                color: '#6b7280',
                            }}
                        >
                            Upload
                        </Button>
                    </Link>
                <ViewMode/>
				<div className="absolute right-0 flex justify-end items-start my-4">
                    <Button onClick={() => onClick('Albums')}className={`${selectedButton === 'Albums' ? 'bg-gray-300' : ''}`}
                        style={{
							backgroundColor: selectedButton === 'Albums' ? '#e5e7eb' : 'transparent',
							color: selectedButton === 'Albums' ? '#1f2937' : '#6b7280',
						}}>
						Albums
					</Button>
                    <Button onClick={() => onClick('Grid')} className={`px-2 ${selectedButton === 'Grid' ? 'bg-gray-300' : ''}`}
                        style={{
							backgroundColor: selectedButton === 'Grid' ? '#e5e7eb' : 'transparent',
							color: selectedButton === 'Grid' ? '#1f2937' : '#6b7280',
						}}>
						Grid
					</Button>
                    <Button onClick={() => onClick('List')} className={`px-4 ${selectedButton === 'List' ? 'bg-gray-300' : ''}`}
                        style={{
							backgroundColor: selectedButton === 'List' ? '#e5e7eb' : 'transparent',
							color: selectedButton === 'List' ? '#1f2937' : '#6b7280',
						}}>
						List
					</Button>
				</div>
                
			</main>
		</div>
	);
}
