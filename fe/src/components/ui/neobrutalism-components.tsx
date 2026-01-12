import React from 'react';

// Reusable Button with Neobrutalist style
export const Button = ({ children, onClick, variant = 'primary', className = '', disabled = false }: any) => {
  const baseStyles = "font-black uppercase text-sm px-10 py-4 border-2 border-black transition-all neo-btn-shadow neo-btn-active disabled:opacity-50 disabled:pointer-events-none";
  const variants = {
    primary: "bg-[#FF8A00] text-black",
    secondary: "bg-white text-black",
    dark: "bg-black text-white"
  };
  
  return (
    <button 
      disabled={disabled}
      onClick={onClick} 
      className={`${baseStyles} ${variants[variant as keyof typeof variants]} ${className}`}
    >
      {children}
    </button>
  );
};

// Reusable Input/Textarea Wrapper
export const FormField = ({ label, children, required }: any) => (
  <div className="space-y-3 w-full">
    <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
      {label} {required && <span className="text-[#FF8A00]">*</span>}
    </label>
    {children}
  </div>
);

// Central Content Container
export const StepContainer = ({ title, children, footer }: any) => (
  <div className="h-full flex flex-col">
    <div className="mb-10">
      <h2 className="text-3xl font-black uppercase tracking-tight mb-2">{title}</h2>
      <div className="h-1 w-20 bg-[#FF8A00] border border-black"></div>
    </div>
    <div className="flex-1">
      {children}
    </div>
    {footer && <div className="pt-12">{footer}</div>}
  </div>
);


