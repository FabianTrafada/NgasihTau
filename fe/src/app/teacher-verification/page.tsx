"use client"

import * as React from "react"
import { motion, AnimatePresence } from "framer-motion"
import { Check, ChevronRight, ChevronLeft, Loader2 } from "lucide-react"
import FileUploader from "@/components/dashboard/assets/FileUploader"
import { cn } from "@/lib/utils"

// --- Types ---
type FormData = {
    firstName: string
    lastName: string
    email: string
    phone: string
    school: string
    experience: string
    address: string
    idCard: File | null
}

// --- Components ---

const Stepper = ({ currentStep }: { currentStep: number }) => {
    const steps = [1, 2, 3]
    return (
        <div className="flex items-center justify-center space-x-4 mb-6">
            {steps.map((step, index) => (
                <React.Fragment key={step}>
                    <div
                        className={cn(
                            "flex h-10 w-10 items-center justify-center rounded-full border-2 border-[#2B2D42] text-sm font-bold transition-colors",
                            step <= currentStep
                                ? "bg-[#FF8811] text-white"
                                : "bg-white text-gray-400"
                        )}
                    >
                        {step}
                    </div>
                    {index < steps.length - 1 && (
                        <div className="h-0.5 w-12 bg-[#2B2D42]" />
                    )}
                </React.Fragment>
            ))}
        </div>
    )
}

const InputField = ({ label, ...props }: React.InputHTMLAttributes<HTMLInputElement> & { label: string }) => (
    <div className="space-y-1">
        <label className="text-sm font-bold text-[#2B2D42] font-[family-name:var(--font-plus-jakarta-sans)]">{label}</label>
        <input
            className="w-full text-[#2B2D42] px-3 py-2 border-[2px] border-gray-300 rounded-md bg-white focus:outline-none transition-all font-[family-name:var(--font-inter)] focus:border-[#FF8811] focus:shadow-[3px_3px_0px_0px_rgba(255,136,17,1)]"
            {...props}
        />
    </div>
)

const Step1 = ({
    data,
    updateData,
}: {
    data: FormData
    updateData: (key: keyof FormData, value: string) => void
}) => (
    <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
            <InputField
                label="First Name"
                placeholder="e.g. Jane"
                value={data.firstName}
                onChange={(e) => updateData("firstName", e.target.value)}
            />
            <InputField
                label="Last Name"
                placeholder="e.g. Doe"
                value={data.lastName}
                onChange={(e) => updateData("lastName", e.target.value)}
            />
        </div>
        <InputField
            label="Email"
            type="email"
            placeholder="jane@example.com"
            value={data.email}
            onChange={(e) => updateData("email", e.target.value)}
        />
        <InputField
            label="Phone"
            placeholder="08xxxx"
            value={data.phone}
            onChange={(e) => updateData("phone", e.target.value)}
        />
    </div>
)

const Step2 = ({
    data,
    updateData,
}: {
    data: FormData
    updateData: (key: keyof FormData, value: string) => void
}) => (
    <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
            <InputField
                label="School / Institution"
                placeholder="e.g. SMAN 1 Jakarta"
                value={data.school}
                onChange={(e) => updateData("school", e.target.value)}
            />
            <InputField
                label="Years Experience"
                type="number"
                placeholder="e.g. 5"
                value={data.experience}
                onChange={(e) => updateData("experience", e.target.value)}
            />
        </div>
        <InputField
            label="Address"
            placeholder="e.g. Pandjaitan street"
            value={data.address}
            onChange={(e) => updateData("address", e.target.value)}
        />
    </div>
)

const Step3 = ({
    data,
    updateData,
}: {
    data: FormData
    updateData: (file: File | null) => void
}) => {
    return (
        <div className="space-y-3">
            <div className="space-y-2">
                <label className="text-sm font-bold text-[#2B2D42]">ID Verification</label>
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
    )
}

// --- Main Page ---

export default function TeacherVerificationPage() {
    const [step, setStep] = React.useState(1)
    const [isLoading, setIsLoading] = React.useState(false)
    const [isSuccess, setIsSuccess] = React.useState(false)
    const [formData, setFormData] = React.useState<FormData>({
        firstName: "",
        lastName: "",
        email: "",
        phone: "",
        school: "",
        experience: "",
        address: "",
        idCard: null,
    })

    const updateField = (key: keyof FormData, value: string) => {
        setFormData((prev) => ({ ...prev, [key]: value }))
    }

    const updateFile = (file: File | null) => {
        setFormData((prev) => ({ ...prev, idCard: file }))
    }

    const handleNext = () => {
        if (step < 3) setStep(step + 1)
    }

    const handleBack = () => {
        if (step > 1) setStep(step - 1)
    }

    const handleSubmit = async () => {
        setIsLoading(true)
        // Simulate API call
        setTimeout(() => {
            setIsLoading(false)
            setIsSuccess(true)
        }, 2000)
    }

    if (isSuccess) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-[#FAF9F6] p-4">
                <div className="w-full max-w-md relative">
                    <div className="absolute top-4 left-4 w-full h-full bg-[#FF8811] rounded-2xl" />
                    <div className="bg-white rounded-2xl p-8 relative border-2 border-[#2B2D42] text-center">
                        <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
                            <Check className="h-8 w-8 text-green-600" />
                        </div>
                        <h3 className="text-2xl font-bold text-[#2B2D42] mb-2">Verification Submitted!</h3>
                        <p className="text-gray-500 mb-6">
                            We have received your details. We will review them and get back to
                            you shortly.
                        </p>
                        <button
                            onClick={() => window.location.href = "/"}
                            className="w-full cursor-pointer bg-[#FF8811] text-white py-3 rounded-lg font-bold border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[4px] active:translate-y-[4px] transition-all"
                        >
                            Back to Home
                        </button>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className="flex min-h-screen flex-col items-center justify-center bg-[#FAF9F6] p-4">
            <div className="mb-6 text-center">
                <h1 className="text-3xl font-bold text-[#2B2D42]">
                    Teacher <span className="text-[#FF8811]">Verification</span>
                </h1>
                <p className="text-gray-500 mt-2">
                    Complete the steps to verify your account
                </p>
            </div>

            <Stepper currentStep={step} />

            <div className="w-full max-w-2xl relative">
                <div className="absolute top-4 left-4 w-full h-full bg-[#FF8811] rounded-2xl" />
                <div className="bg-white rounded-2xl p-6 relative border-2 border-[#2B2D42] overflow-hidden">
                    <div className="relative min-h-[300px]">
                        <AnimatePresence mode="wait">
                            <motion.div
                                key={step}
                                initial={{ x: 20, opacity: 0 }}
                                animate={{ x: 0, opacity: 1 }}
                                exit={{ x: -20, opacity: 0 }}
                                transition={{ duration: 0.2 }}
                            >
                                <div className="mb-6">
                                    <h3 className="text-xl font-bold text-[#2B2D42]">
                                        {step === 1 && "Personal Information"}
                                        {step === 2 && "Professional Details"}
                                        {step === 3 && "ID Verification"}
                                    </h3>
                                    <p className="text-sm text-gray-500">
                                        {step === 1 && "Tell us a bit about yourself."}
                                        {step === 2 && "Where do you teach?"}
                                        {step === 3 && "Upload a valid ID card for verification."}
                                    </p>
                                </div>

                                {step === 1 && (
                                    <Step1 data={formData} updateData={updateField} />
                                )}
                                {step === 2 && (
                                    <Step2 data={formData} updateData={updateField} />
                                )}
                                {step === 3 && (
                                    <Step3 data={formData} updateData={updateFile} />
                                )}
                            </motion.div>
                        </AnimatePresence>
                    </div>

                    <div className="flex justify-between mt-8 pt-4 border-t border-gray-100">
                        <button
                            onClick={handleBack}
                            disabled={step === 1 || isLoading}
                            className={cn(
                                "flex items-center px-4 py-2 text-sm font-bold text-[#2B2D42] hover:text-[#FF8811] transition-colors disabled:opacity-50",
                                step === 1 ? "invisible" : ""
                            )}
                        >
                            <ChevronLeft className="mr-2 h-4 w-4" /> Back
                        </button>

                        {step < 3 ? (
                            <button
                                onClick={handleNext}
                                className="cursor-pointer flex items-center px-6 py-2 bg-[#FF8811] text-white rounded-lg font-bold border-2 border-[#2B2D42] shadow-[3px_3px_0px_0px_#2B2D42] hover:translate-x-[1px] hover:translate-y-[1px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[3px] active:translate-y-[3px] transition-all"
                            >
                                Next <ChevronRight className="ml-2 h-4 w-4" />
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
                                        Verifying...
                                    </>
                                ) : (
                                    "Verify Now"
                                )}
                            </button>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}
