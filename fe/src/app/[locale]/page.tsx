"use client";

import { Navbar } from "@/components/landing-page/navbar";
import { Hero } from "@/components/landing-page/hero";
import { Features } from "@/components/landing-page/features";
import { SearchSection } from "@/components/landing-page/search-section";
import { KnowledgePods } from "@/components/landing-page/knowledge-pods";
import { Testimonials } from "@/components/landing-page/testimonials";
import { Footer } from "@/components/landing-page/footer";
import { getUserLearningStatus } from "@/lib/api/user";
import { getUserBehavior } from "@/lib/api/behavior";
import { useEffect, useState } from "react";
import { UserBehavior } from "@/types/userBehavior";

export default function Home() {
  const [behaviorData, setBehaviorData] = useState<UserBehavior | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchBehaviorData = async () => {
      try {
        setIsLoading(true);
        const data = await getUserBehavior();
        setBehaviorData(data);
        
        // Get learning status based on behavior data
        const learningStatus = await getUserLearningStatus(data);
        console.log("Learning Status:", learningStatus);
      } catch (error) {
        console.error("Error fetching behavior data:", error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchBehaviorData();
  }, []);

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
