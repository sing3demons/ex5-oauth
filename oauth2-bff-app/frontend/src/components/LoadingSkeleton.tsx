export default function LoadingSkeleton() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-500 to-purple-600 p-4 md:p-6 lg:p-8">
      {/* Header Skeleton */}
      <div className="flex flex-col sm:flex-row justify-between items-center p-3 md:p-4 lg:p-6 bg-white/95 rounded-xl mb-6 shadow-md gap-3 sm:gap-0 animate-pulse">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-gray-300 rounded"></div>
          <div className="w-24 h-6 bg-gray-300 rounded"></div>
        </div>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-gray-300 rounded-full"></div>
          <div className="hidden sm:flex flex-col gap-2">
            <div className="w-24 h-4 bg-gray-300 rounded"></div>
            <div className="w-32 h-3 bg-gray-300 rounded"></div>
          </div>
        </div>
      </div>

      {/* Page Title Skeleton */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-6 md:mb-8 gap-4 animate-pulse">
        <div>
          <div className="w-48 h-10 bg-white/30 rounded mb-2"></div>
          <div className="w-64 h-5 bg-white/20 rounded"></div>
        </div>
        <div className="w-32 h-12 bg-white/30 rounded-lg"></div>
      </div>

      {/* Stats Skeleton */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 md:gap-6 mb-6 md:mb-8 animate-pulse">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-white/95 rounded-xl p-4 md:p-6 text-center shadow-md">
            <div className="w-16 h-12 bg-gray-300 rounded mx-auto mb-2"></div>
            <div className="w-20 h-4 bg-gray-300 rounded mx-auto"></div>
          </div>
        ))}
      </div>

      {/* Board Skeleton */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6 items-start">
        {[1, 2, 3].map((col) => (
          <div key={col} className="bg-white/95 rounded-xl p-4 md:p-5 lg:p-6 min-h-[400px] shadow-md animate-pulse">
            <div className="flex justify-between items-center mb-4 pb-3 border-b-2 border-gray-200">
              <div className="flex items-center gap-3">
                <div className="w-24 h-6 bg-gray-300 rounded"></div>
                <div className="w-8 h-6 bg-gray-300 rounded-xl"></div>
              </div>
            </div>
            <div className="space-y-3">
              {[1, 2, 3].map((card) => (
                <div key={card} className="bg-gray-100 rounded-lg p-3 md:p-4">
                  <div className="flex justify-between items-center mb-3">
                    <div className="w-12 h-5 bg-gray-300 rounded"></div>
                    <div className="flex gap-2">
                      <div className="w-8 h-8 bg-gray-300 rounded"></div>
                      <div className="w-8 h-8 bg-gray-300 rounded"></div>
                    </div>
                  </div>
                  <div className="w-full h-5 bg-gray-300 rounded mb-2"></div>
                  <div className="w-3/4 h-4 bg-gray-300 rounded mb-3"></div>
                  <div className="w-20 h-3 bg-gray-300 rounded"></div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
