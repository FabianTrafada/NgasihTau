"use client";

import { motion } from "framer-motion";
import { fadeInUp, staggerContainer } from "@/lib/animations";

export function LearnCompete() {
    return (
        <section className="py-20 px-6 md:px-12 max-w-7xl mx-auto w-full relative z-10">
            <div className="flex flex-col md:flex-row items-center justify-center gap-16">
                {/* Icon */}
                <motion.div
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true }}
                    variants={staggerContainer}
                    className="flex items-end gap-4"
                >
                    <motion.div variants={fadeInUp} className="flex flex-col items-center gap-2">
                        <div className="w-8 h-8 rounded-full bg-[#FF8811]"></div>
                        <div className="w-12 h-24 bg-[#FF8811] rounded-t-md"></div>
                    </motion.div>
                    <motion.div variants={fadeInUp} className="flex flex-col items-center gap-2">
                        <div className="w-8 h-8 rounded-full bg-[#FF8811]"></div>
                        <div className="w-12 h-32 bg-[#FF8811] rounded-t-md"></div>
                    </motion.div>
                    <motion.div variants={fadeInUp} className="flex flex-col items-center gap-2">
                        <div className="w-8 h-8 rounded-full bg-[#FF8811]"></div>
                        <div className="w-12 h-28 bg-[#FF8811] rounded-t-md"></div>
                    </motion.div>
                </motion.div>

                {/* Text */}
                <motion.div
                    initial={{ opacity: 0, x: 50 }}
                    whileInView={{ opacity: 1, x: 0 }}
                    viewport={{ once: true }}
                    transition={{ duration: 0.8 }}
                    className="max-w-lg"
                >
                    <h2 className="text-4xl font-bold text-[#2B2D42] mb-6">
                        Learn. Compete. <br />
                        Contribute.
                    </h2>
                    <p className="text-gray-600 font-[family-name:var(--font-inter)] text-lg leading-relaxed">
                        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.
                    </p>
                </motion.div>
            </div>
        </section>
    );
}
