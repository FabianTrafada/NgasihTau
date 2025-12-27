"use client";

import { motion } from "framer-motion";
import { fadeInUp, staggerContainer, scaleIn } from "@/lib/animations";
import { features } from "@/lib/data/landing-page";
import { Layers, Bot, Users } from "lucide-react";

const iconMap = {
    Layers: Layers,
    Bot: Bot,
    Users: Users,
};

export function Features() {
    return (
        <section className="py-20 px-6 md:px-12 max-w-7xl mx-auto w-full relative z-10" id="features">
            <motion.h2
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true }}
                variants={fadeInUp}
                className="text-4xl font-bold text-center text-[#2B2D42] mb-16"
            >
                Features
            </motion.h2>
            <motion.div
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true }}
                variants={staggerContainer}
                className="grid grid-cols-1 md:grid-cols-3 gap-8"
            >
                {features.map((feature, index) => {
                    const Icon = iconMap[feature.icon as keyof typeof iconMap];
                    return (
                        <motion.div
                            key={index}
                            variants={scaleIn}
                            className={`bg-white p-8 rounded-xl border-2 border-[#2B2D42] ${feature.shadowColor} shadow-[8px_8px_0px_0px] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[4px_4px_0px_0px] transition-all h-full flex flex-col`}
                        >
                            <div className="w-12 h-12 bg-gray-100 rounded-md mb-6 flex items-center justify-center border-2 border-[#2B2D42]">
                                <Icon className="w-6 h-6 text-[#2B2D42]" />
                            </div>
                            <h3 className="text-xl font-bold text-[#2B2D42] mb-4">{feature.title}</h3>
                            <p className="text-gray-600 font-[family-name:var(--font-inter)] leading-relaxed">
                                {feature.description}
                            </p>
                        </motion.div>
                    );
                })}
            </motion.div>
        </section>
    );
}
