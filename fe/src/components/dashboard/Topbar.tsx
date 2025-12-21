import { Bell, Search } from 'lucide-react'
import React from 'react'

const Topbar = () => {
    return (
        <header className='h-20 px-8 flex items-center justify-between bg-[#fffbf7] sticky top-0 z-10'>

            {/* searchbar */}
            <div className='relative w-96'>
                <Search className='absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400' />

                <input type="text"
                    placeholder='search'
                    className='w-full pl-10 pr-4 py-2.5 bg-gray-100/50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#ff8811]/20 focus:border-[#ff8811] transition-all  ' />
            </div>

            {/* right actions */}

            <div className='flex items-center gap-6'>
                <button className="relative p-2 text-gray-600 hover:text-[#FF8811] transition-colors">
                    <Bell className="w-6 h-6" />
                    <span className="absolute top-1.5 right-2 w-2 h-2 bg-red-500 rounded-full border-2 border-[#FFFBF7]"></span>
                </button>

                <div className="w-10 h-10 rounded-full bg-[#FF8811] flex items-center justify-center text-white font-bold shadow-sm cursor-pointer hover:opacity-90 transition-opacity">
                    TL
                </div>

            </div>

        </header>
    )
}

export default Topbar