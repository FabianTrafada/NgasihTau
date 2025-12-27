"use client";

import { motion } from "framer-motion";
import { Marquee } from "@/components/ui/marquee";
import { cn } from "@/lib/utils";
import { testimonials } from "@/lib/data/landing-page";

const ReviewCard = ({
    img,
    name,
    username,
    body,
}: {
    img: string;
    name: string;
    username: string;
    body: string;
}) => {
    return (
        <figure
            className={cn(
                "relative w-80 cursor-pointer overflow-hidden rounded-xl border-2 border-[#2B2D42] p-6",
                "bg-[#FFFBF7] hover:bg-[#FFFBF7]/90 transition-colors",
                "shadow-[4px_4px_0px_0px_#2B2D42]"
            )}
        >
            <div className="flex flex-row items-center gap-2" id="testimonials">
                <img className="rounded-full" width="32" height="32" alt="" src={img} />
                <div className="flex flex-col">
                    <figcaption className="text-sm font-bold text-[#2B2D42]">
                        {name}
                    </figcaption>
                    <p className="text-xs font-medium text-gray-500">{username}</p>
                </div>
            </div>
            <blockquote className="mt-4 text-sm text-[#2B2D42] font-[family-name:var(--font-inter)]">{body}</blockquote>
        </figure>
    );
};

export function Testimonials() {
    return (
        <section className="py-20 w-full relative z-10 overflow-hidden">
            <motion.h2
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6 }}
                className="text-4xl font-bold text-center text-[#2B2D42] mb-16"
            >
                Testimonials
            </motion.h2>
            <div className="relative flex w-full flex-col items-center justify-center overflow-hidden">
                <Marquee pauseOnHover className="[--duration:40s]">
                    {testimonials.map((review, i) => (
                        <ReviewCard key={i} {...review} />
                    ))}
                </Marquee>
                <Marquee reverse pauseOnHover className="[--duration:40s] mt-8">
                    {testimonials.map((review, i) => (
                        <ReviewCard key={i} {...review} />
                    ))}
                </Marquee>
                <div className="pointer-events-none absolute inset-y-0 left-0 w-1/3 bg-gradient-to-r from-[#FFFBF7] dark:from-background"></div>
                <div className="pointer-events-none absolute inset-y-0 right-0 w-1/3 bg-gradient-to-l from-[#FFFBF7] dark:from-background"></div>
            </div>
        </section>
    );
}
