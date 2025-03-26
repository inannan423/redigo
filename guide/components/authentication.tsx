"use client"
import { useState } from 'react';
import Cookies from 'js-cookie';
import { IconKey, IconCheck } from '@tabler/icons-react';

const Authentication = () => {
  const [token, setToken] = useState('');
  const [tokenType, setTokenType] = useState<'admin' | 'api'>('admin');
  const [showSuccess, setShowSuccess] = useState(false);

  const handleTokenSave = () => {
    if (!token) return;
    
    const cookieName = tokenType === 'admin' ? 'adminToken' : 'apiToken';
    Cookies.set(cookieName, token, { expires: 7 });
    setToken('');
    setShowSuccess(true);
    setTimeout(() => setShowSuccess(false), 3000);
  };

  return (
    <div className="w-full my-5 rounded-lg border flex flex-col border-gray-200 bg-white dark:bg-gray-800 dark:border-gray-700 h-max p-4">
      <div className="w-full flex flex-col mb-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-100 dark:border-gray-700 px-4 py-3 h-max">
        <div className="text-base font-bold text-gray-700 dark:text-gray-300 flex items-center mb-3">
          <IconKey size={18} className="mr-2 text-gray-600 dark:text-gray-400" />
          <span>Authentication Token</span>
        </div>

        <div className="space-y-3">
          <input
            type="text"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="Enter your token"
            className="w-full px-3 py-2 text-sm font-mono bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md focus:outline-none focus:ring-1 focus:ring-slate-400 dark:focus:ring-slate-500 focus:border-slate-400 dark:focus:border-slate-500 transition-all"
          />
          
          <div className="flex space-x-4">
            <label className="flex items-center">
              <input
                type="radio"
                checked={tokenType === 'admin'}
                onChange={() => setTokenType('admin')}
                className="mr-2 text-slate-600"
              />
              <span className="text-sm text-gray-700 dark:text-gray-300">Admin Token</span>
            </label>
            
            <label className="flex items-center">
              <input
                type="radio"
                checked={tokenType === 'api'}
                onChange={() => setTokenType('api')}
                className="mr-2 text-slate-600"
              />
              <span className="text-sm text-gray-700 dark:text-gray-300">API Token</span>
            </label>
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        {showSuccess && (
          <div className="flex items-center gap-2 p-3 bg-green-50 dark:bg-green-900/30 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-400 rounded-lg">
            <IconCheck size={18} className="flex-shrink-0" />
            <span className="text-sm">Token saved successfully!</span>
          </div>
        )}

        <button
          onClick={handleTokenSave}
          className="w-full bg-slate-600 hover:bg-slate-500 text-sm transition-all duration-300 text-white font-mono px-3 py-2 rounded-lg cursor-pointer flex justify-center items-center"
        >
          <IconKey size={18} className="mr-2" />
          Save Token
        </button>
      </div>
    </div>
  );
};

export default Authentication;