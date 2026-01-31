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
import { getCurrentUser } from "@/lib/api/auth";
import { getBehaviorData, fetchUserPersona } from '@/lib/api/behavior';

export default function Home() {
  // const behaviorData: UserBehavior = {
  //   user_id: "user-123",
  //   behavior_data: {
  //     user_id: "user-123",
  //     analysis_period_days: 30,
  //     chat: {
  //       total_messages: 150,
  //       user_messages: 80,
  //       assistant_messages: 70,
  //       question_count: 45,
  //       avg_message_length: 120.5,
  //       thumbs_up_count: 25,
  //       thumbs_down_count: 5,
  //       unique_sessions: 20,
  //       total_session_duration_minutes: 180.0,
  //     },
  //     material: {
  //       total_time_spent_seconds: 200,
  //       total_views: 5,
  //       unique_materials_viewed: 12,
  //       bookmark_count: 8,
  //       avg_scroll_depth: 0.15,
  //     },
  //     activity: {
  //       active_days: 18,
  //       total_sessions: 35,
  //       peak_hour: 14,
  //       late_night_sessions: 3,
  //       weekend_sessions: 8,
  //       total_weekday_sessions: 27,
  //       daily_activity_variance: 2.5,
  //     },
  //     quiz: {
  //       quiz_attempts: 100,
  //       avg_score: 18.5,
  //       completion_rate: 0.85,
  //     },
  //   },
  //   quiz_score: 18.5,
  //   previous_persona: 'false',
  // };
  // useEffect(() => {
  //   const fetchUserLearningStatus = async () => {
  //     const response = await getUserLearningStatus(behaviorData);
  //     console.log(response);
  //     return response;
  //   };

    // const fetchCurrentUser = async () => {
    //     const response = await getCurrentUser();
    //     console.log("Current User:", response);
    //     return response;
    // }
    // fetchUserLearningStatus();
    // fetchCurrentUser();
  // }, []);

  // Test get behavior data
  useEffect(() => {
    const fetch = async () => {
      const behavior = await getBehaviorData();
      console.log('Behavior:', behavior);

      // Test full persona prediction
      const persona = await fetchUserPersona('e409b45f-97bc-4b4a-9029-547650c9f9e3');
      console.log('Persona:', persona);
    }

    fetch();
  })

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
