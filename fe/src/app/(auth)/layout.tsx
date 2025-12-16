export default function AuthLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <div className="relative min-h-screen bg-[#FFFBF7] flex flex-col">
            {/* Background pattern */}
            <div
                className="absolute inset-0"
                style={{
                    backgroundImage: `radial-gradient(circle, #D4D4D4 1px, transparent 1px)`,
                    backgroundSize: "24px 24px",
                    opacity: 0.6,
                }}
            />

            {/* Centered content */}
            <main className="relative z-10 flex flex-1 w-full items-center justify-center px-4 py-12">
                {children}
            </main>

            <footer className="relative z-10 py-4 text-center">
                <p className="text-gray-400 text-sm font-[family-name:var(--font-inter)]">
                    © 2025 NgasihTau. All rights reserved.
                </p>
            </footer>
        </div>
    );
}
