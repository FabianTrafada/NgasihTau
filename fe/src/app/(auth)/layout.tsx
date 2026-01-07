export default function AuthLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <div className="relative min-h-screen bg-[#FFFBF7] flex flex-col gap-3">
            {/* Background pattern */}
            <div
                className="absolute inset-0"
                style={{
                    backgroundImage: `radial-gradient(circle, #D4D4D4 1px, transparent 1px)`,
                    backgroundSize: "24px 24px",
                    opacity: 0.6,
                }}
            />

            {/* svg content */}
            <div className="relative z-10 flex flex-1 w-full max-w-full mx-auto items-center">
                <main className="flex flex-[1.2]  items-center justify-center px-4 py-12 lg:justify-start lg:pl-20">
                    <div className="w-full max-w-lg">
                        {children}
                    </div>
                </main>
                <div className="hidden lg:flex flex-1 items-center justify-start p-12">
                    <div className="w-full max-w-md bg-white/50 backdrop-blur-sm p-8 rounded-3xl border border-gray-100 shadow-sm">
                        <svg
                            viewBox="0 0 500 500"
                            className="w-full max-w-md h-auto"
                            xmlns="http://www.w3.org/2000/svg"
                        >
                            {/* Contoh SVG Geometris agar tidak terlalu clean */}
                            <defs>
                                <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="100%">
                                    <stop offset="0%" style={{ stopColor: '#FF6B6B', stopOpacity: 1 }} />
                                    <stop offset="100%" style={{ stopColor: '#4ECDC4', stopOpacity: 1 }} />
                                </linearGradient>
                            </defs>
                            <circle cx="250" cy="250" r="200" fill="url(#grad1)" opacity="0.1" />
                            <rect x="100" y="100" width="300" height="300" rx="20" fill="currentColor" className="text-orange-500/10" transform="rotate(15 250 250)" />
                            <path d="M150 150 L350 350 M350 150 L150 350" stroke="currentColor" strokeWidth="2" className="text-gray-300" />
                            {/* Anda bisa mengganti ini dengan <img src="/your-svg.svg" /> jika sudah ada filenya */}
                        </svg>

                    </div>
                </div>
            </div>

            <footer className="relative z-10 py-4 text-center">
                <p className="text-gray-400 text-sm font-[family-name:var(--font-inter)]">
                    © 2025 NgasihTau. All rights reserved.
                </p>
            </footer>
        </div>
    );
}
