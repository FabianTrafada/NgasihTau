"use client";

import * as React from "react";
import { Check, ChevronRight, ChevronLeft, Loader2 } from "lucide-react";
import FileUploader from "@/components/dashboard/assets/FileUploader";
import { cn } from "@/lib/utils";
import { useTranslations } from "next-intl";

// --- Types ---
type FormData = {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  school: string;
  experience: string;
  address: string;
  idCard: File | null;
};

// --- Components ---

const InputField = ({ label, ...props }: React.InputHTMLAttributes<HTMLInputElement> & { label: string }) => (
  <div className="space-y-1">
    <label className="text-sm font-bold text-[#2B2D42] font-[family-name:var(--font-plus-jakarta-sans)]">{label}</label>
    <input
      className="w-full text-[#2B2D42] px-3 py-2 border-[2px] border-gray-300 rounded-md bg-white focus:outline-none transition-all font-[family-name:var(--font-inter)] focus:border-[#FF8811] focus:shadow-[3px_3px_0px_0px_rgba(255,136,17,1)] disabled:opacity-50 disabled:cursor-not-allowed"
      {...props}
    />
  </div>
);

const Step1 = ({ data, updateData, disabled }: { data: FormData; updateData: (key: keyof FormData, value: string) => void; disabled?: boolean }) => {
  const t = useTranslations("teacherVerification");
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <InputField label={t("firstName")} placeholder="e.g. Jane" value={data.firstName} onChange={(e) => updateData("firstName", e.target.value)} disabled={disabled} />
        <InputField label={t("lastName")} placeholder="e.g. Doe" value={data.lastName} onChange={(e) => updateData("lastName", e.target.value)} disabled={disabled} />
      </div>
      <InputField label={t("email")} type="email" placeholder="jane@example.com" value={data.email} onChange={(e) => updateData("email", e.target.value)} disabled={disabled} />
      <InputField label={t("phone")} placeholder="08xxxx" value={data.phone} onChange={(e) => updateData("phone", e.target.value)} disabled={disabled} />
    </div>
  );
};

const Step2 = ({ data, updateData, disabled }: { data: FormData; updateData: (key: keyof FormData, value: string) => void; disabled?: boolean }) => {
  const t = useTranslations("teacherVerification");
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <InputField label={t("school")} placeholder="e.g. SMAN 1 Jakarta" value={data.school} onChange={(e) => updateData("school", e.target.value)} disabled={disabled} />
        <InputField label={t("experience")} type="number" placeholder="e.g. 5" value={data.experience} onChange={(e) => updateData("experience", e.target.value)} disabled={disabled} />
      </div>
      <InputField label={t("address")} placeholder="e.g. Pandjaitan street" value={data.address} onChange={(e) => updateData("address", e.target.value)} disabled={disabled} />
    </div>
  );
};

const Step3 = ({ data, updateData, disabled }: { data: FormData; updateData: (file: File | null) => void; disabled?: boolean }) => {
  return (
    <div className="space-y-3">
      <div className="space-y-2">
        <label className="text-sm font-bold text-[#2B2D42]">ID Verification</label>
        <div className={disabled ? "pointer-events-none opacity-50" : ""}>
          <FileUploader
            onSingleFileSelect={updateData}
            accept={{ "image/*": [".jpg", ".jpeg", ".png", ".svg", ".gif"] }}
            maxSize={5 * 1024 * 1024}
            multiple={false}
            label="Click here"
            description="Supported formats: JPG, PNG, SVG (Max 5MB)"
            selectedFile={data.idCard}
          />
        </div>
      </div>
    </div>
  );
};

// --- Main Page ---

export default function TeacherVerificationPage() {
  const t = useTranslations("teacherVerification");
  const [step, setStep] = React.useState(1);
  const [isLoading, setIsLoading] = React.useState(false);
  const [isSuccess, setIsSuccess] = React.useState(false);
  const [formData, setFormData] = React.useState<FormData>({
    firstName: "",
    lastName: "",
    email: "",
    phone: "",
    school: "",
    experience: "",
    address: "",
    idCard: null,
  });

  const stepRefs = React.useRef<(HTMLDivElement | null)[]>([]);

  React.useEffect(() => {
    if (stepRefs.current[step - 1]) {
      stepRefs.current[step - 1]?.scrollIntoView({
        behavior: "smooth",
        block: "center",
      });
    }
  }, [step]);

  const updateField = (key: keyof FormData, value: string) => {
    setFormData((prev) => ({ ...prev, [key]: value }));
  };

  const updateFile = (file: File | null) => {
    setFormData((prev) => ({ ...prev, idCard: file }));
  };

  const handleNext = () => {
    if (step < 3) setStep(step + 1);
  };

  const handleBack = () => {
    if (step > 1) setStep(step - 1);
  };

  const handleSubmit = async () => {
    setIsLoading(true);
    // Simulate API call
    setTimeout(() => {
      setIsLoading(false);
      setIsSuccess(true);
    }, 2000);
  };

  const steps = [
    { id: 1, title: t("step1"), description: t("subtitle") },
    { id: 2, title: t("step2"), description: t("subtitle") },
    { id: 3, title: t("step3"), description: t("subtitle") },
  ];

  if (isSuccess) {
    return (
      <div className="flex h-screen w-full items-center justify-center bg-[#FAF9F6] p-4">
        <div className="w-full max-w-md relative">
          <div className="absolute top-4 left-4 w-full h-full bg-[#FF8811] rounded-2xl" />
          <div className="bg-white rounded-2xl p-8 relative border-2 border-[#2B2D42] text-center">
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <Check className="h-8 w-8 text-green-600" />
            </div>
            <h3 className="text-2xl font-bold text-[#2B2D42] mb-2">{t("success")}</h3>
            <p className="text-gray-500 mb-6">{t("successMessage")}</p>
            <button
              onClick={() => (window.location.href = "/")}
              className="w-full cursor-pointer bg-[#FF8811] text-white py-3 rounded-lg font-bold border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[4px] active:translate-y-[4px] transition-all"
            >
              {t("backToDashboard")}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen w-full flex-col items-center bg-[#FAF9F6] overflow-hidden">
      <div className="w-full max-w-3xl shrink-0 pt-10 pb-6 text-center px-4">
        <h1 className="text-3xl font-bold text-[#2B2D42]">
          {t("title").split(" ")[0]} <span className="text-[#FF8811]">{t("title").split(" ")[1]}</span>
        </h1>
        <p className="text-gray-500 mt-2">{t("subtitle")}</p>
      </div>

      <div className="w-full max-w-3xl flex-1 overflow-y-auto px-4 pb-10 scrollbar-hide">
        <div className="relative">
          {/* Vertical Line */}
          <div className="absolute left-[19px] top-4 bottom-0 w-0.5 bg-[#2B2D42] -z-10" />

          <div className="flex flex-col gap-8">
            {steps.map((s, index) => {
              const isActive = step === s.id;
              const isCompleted = step > s.id;

              return (
                <div
                  key={s.id}
                  className="flex gap-6"
                  ref={(el) => {
                    stepRefs.current[index] = el;
                  }}
                >
                  {/* Indicator */}
                  <div className="flex-shrink-0">
                    <div
                      className={cn(
                        "flex h-10 w-10 items-center justify-center rounded-full border-2 border-[#2B2D42] text-sm font-bold transition-colors z-10 relative",
                        isActive || isCompleted ? "bg-[#FF8811] text-white" : "bg-white text-gray-400",
                      )}
                    >
                      {isCompleted ? <Check className="h-5 w-5" /> : s.id}
                    </div>
                  </div>

                  {/* Content */}
                  <div
                    className={cn(
                      "flex-grow transition-all duration-300 rounded-xl p-6 border-2",
                      isActive ? "bg-white border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] opacity-100" : "bg-transparent border-transparent shadow-none opacity-60",
                    )}
                  >
                    <div className="mb-6">
                      <h3 className="text-xl font-bold text-[#2B2D42]">{s.title}</h3>
                      <p className="text-sm text-[#2B2D42]/80">{s.description}</p>
                    </div>

                    <div className={cn("transition-all", !isActive && "pointer-events-none")}>
                      {s.id === 1 && <Step1 data={formData} updateData={updateField} disabled={!isActive} />}
                      {s.id === 2 && <Step2 data={formData} updateData={updateField} disabled={!isActive} />}
                      {s.id === 3 && <Step3 data={formData} updateData={updateFile} disabled={!isActive} />}
                    </div>

                    {isActive && (
                      <div className="flex justify-between mt-8 pt-4 border-t border-[#2B2D42]/20">
                        <button
                          onClick={handleBack}
                          disabled={step === 1 || isLoading}
                          className={cn("flex items-center px-4 py-2 text-sm font-bold text-[#2B2D42] hover:text-white transition-colors disabled:opacity-50", step === 1 ? "invisible" : "")}
                        >
                          <ChevronLeft className="mr-2 h-4 w-4" /> {t("previous")}
                        </button>

                        {step < 3 ? (
                          <button
                            onClick={handleNext}
                            className="cursor-pointer flex items-center px-6 py-2 bg-[#FF8811] text-white rounded-lg font-bold border-2 border-[#2B2D42] shadow-[3px_3px_0px_0px_#2B2D42] hover:translate-x-[1px] hover:translate-y-[1px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[3px] active:translate-y-[3px] transition-all"
                          >
                            {t("next")} <ChevronRight className="ml-2 h-4 w-4" />
                          </button>
                        ) : (
                          <button
                            onClick={handleSubmit}
                            disabled={isLoading || !formData.idCard}
                            className="cursor-pointer flex items-center px-6 py-2 bg-[#FF8811] text-white rounded-lg font-bold border-2 border-[#2B2D42] shadow-[3px_3px_0px_0px_#2B2D42] hover:translate-x-[1px] hover:translate-y-[1px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[3px] active:translate-y-[3px] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                          >
                            {isLoading ? (
                              <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                {t("submitting")}
                              </>
                            ) : (
                              t("submit")
                            )}
                          </button>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}