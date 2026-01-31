"use client";

import { Navbar } from "@/components/landing-page/navbar";
import { Hero } from "@/components/landing-page/hero";
import { Features } from "@/components/landing-page/features";
import { SearchSection } from "@/components/landing-page/search-section";
import { KnowledgePods } from "@/components/landing-page/knowledge-pods";
import { Testimonials } from "@/components/landing-page/testimonials";
import { Footer } from "@/components/landing-page/footer";
import { getUserLearningStatus } from "@/lib/api/user";
import { useEffect, useState } from "react";
import { UserBehavior } from "@/types/userBehavior";
import { getCurrentUser } from "@/lib/api/auth";
import { getBehaviorData, fetchUserPersona } from "@/lib/api/behavior";

export default function Home() {
  return (
    <div className="min-h-screen bg-[#FFFBF7] font-family-name:var(--font-plus-jakarta-sans) overflow-x-hidden relative flex flex-col">
      <Navbar />
      <Hero />
      <Features />
      <SearchSection />
      <KnowledgePods />
      <Testimonials />
      <Footer />
    </div>
  );
}
