"use client"
import React, { useState, useEffect } from 'react';
import { IconSend, IconPlus, IconX, IconCode, IconCloudDownload, IconClock, IconAlertCircle } from '@tabler/icons-react';
import Cookies from 'js-cookie';

type RequesterProps = {
    method?: "GET" | "POST" | "PUT" | "DELETE" | "PATCH",
    url?: string,
    defaultHeaders?: Array<{ key: string; value: string }>,
    defaultBody?: string,
    description?: string,
    type?: 'admin' | 'api',
    isMultipart?: boolean 
}

function MethodColor(method: string) {
    switch (method) {
        case "GET":
            return "text-green-500";
        case "POST":
            return "text-blue-500";
        case "PUT":
            return "text-yellow-500";
        case "DELETE":
            return "text-red-500";
        case "PATCH":
            return "text-purple-500";
        default:
            return "text-gray-500";
    }
}

export default function Requester({
    method = "POST",
    url = "/api",
    defaultHeaders = [{ key: 'Content-Type', value: 'application/json' }],
    defaultBody = '',
    description = '',
    type = 'admin',
    isMultipart = false
}: RequesterProps) {
    const [headers, setHeaders] = useState<Array<{ key: string; value: string }>>(defaultHeaders);
    const [body, setBody] = useState<string>(defaultBody);
    const [response, setResponse] = useState<null | {
        status: number,
        statusText: string,
        data: any,
        time: number
    }>(null);
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);
    const [urlParams, setUrlParams] = useState<Record<string, string>>({});
    const baseUrl = 'http://localhost:8080';

    const handleBodyChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setBody(e.target.value);
    };

    const addHeader = () => {
        setHeaders([...headers, { key: '', value: '' }]);
    };

    const removeHeader = (index: number) => {
        setHeaders(headers.filter((_, i) => i !== index));
    };

    const updateHeader = (index: number, field: 'key' | 'value', value: string) => {
        const newHeaders = [...headers];
        newHeaders[index][field] = value;
        setHeaders(newHeaders);
    };
    // Extract URL parameters from the url prop
    useEffect(() => {
        const paramRegex = /:(\w+)/g;
        const matches = url.match(paramRegex);
        if (matches) {
            const initialParams: Record<string, string> = {};
            matches.forEach(match => {
                const paramName = match.slice(1); // Remove the : prefix
                initialParams[paramName] = '';
            });
            setUrlParams(initialParams);
        }
    }, [url]);

    // Construct URL with parameters
    const getFullUrl = () => {
        let processedUrl = url;
        Object.entries(urlParams).forEach(([key, value]) => {
            processedUrl = processedUrl.replace(`:${key}`, value);
        });
        return `${baseUrl}${processedUrl}`;
    };

    const [file, setFile] = useState<File | null>(null);  // 新增状态

    const sendRequest = async () => {
        setLoading(true);
        setError(null);
        setResponse(null);
        
        const startTime = performance.now();
        
        try {
            const token = Cookies.get(type === 'admin' ? 'adminToken' : 'apiToken');
            
            if (!token) {
                setError(`No ${type} token found. Please set your token first.`);
                setLoading(false);
                return;
            }
            
            // Prepare headers
            const headerObj: Record<string, string> = {};
            headers.forEach(h => {
                if (h.key.trim()) {
                    headerObj[h.key] = h.value;
                }
            });
            
            headerObj['Authorization'] = `Bearer ${token}`;
            
            // Prepare request options
            const options: RequestInit = {
                method,
                headers: headerObj,
            };
            
            if (method !== 'GET') {
                if (isMultipart && file) {
                    const formData = new FormData();
                    formData.append('file', file);
                    options.body = formData;
                    delete headerObj['Content-Type'];
                } else if (body.trim()) {
                    try {
                        const jsonBody = JSON.parse(body);
                        options.body = JSON.stringify(jsonBody);
                    } catch (e) {
                        options.body = body;
                    }
                }
            }
            
            const fullUrl = getFullUrl();
            
            // Send request
            const res = await fetch(fullUrl, options);
            const endTime = performance.now();
            
            // Try to parse response as JSON
            let data;
            const contentType = res.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                data = await res.json();
            } else {
                data = await res.text();
            }
            
            // Set response
            setResponse({
                status: res.status,
                statusText: res.statusText,
                data,
                time: Math.round(endTime - startTime)
            });
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An unknown error occurred');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="w-full my-5 rounded-lg border flex flex-col border-gray-200 bg-white dark:bg-gray-800 dark:border-gray-700 h-max p-4">
            {description && (
                <div className="mb-3 text-sm text-gray-600 dark:text-gray-300">
                    {description}
                </div>
            )}
            
            <div className="w-full flex flex-col gap-2 mb-2">
                <div className='bg-gray-50 dark:bg-gray-900 border border-gray-100 dark:border-gray-700 w-full px-4 py-2 rounded-lg flex gap-5 justify-start items-center'>
                    <div className={`px-2 py-1 rounded font-mono text-sm font-bold ${MethodColor(method)}`}>
                        {method}
                    </div>
                    <div className="font-mono text-sm text-gray-500 dark:text-gray-400 flex items-center gap-2">
                        <span className="font-semibold">Base URL</span>
                    </div>
                    <div className='h-full border-r border-gray-300 dark:border-gray-600'></div>
                    <div className='font-mono text-sm italic text-gray-500 dark:text-gray-400'>
                        {url}
                    </div>
                </div>
            
            {/* URL Parameters Section */}
            {Object.keys(urlParams).length > 0 && (
                <div className='bg-gray-50 dark:bg-gray-900 border border-gray-100 dark:border-gray-700 rounded-lg p-3'>
                    <div className='text-sm font-bold text-gray-700 dark:text-gray-300 mb-2'>
                        URL Parameters
                    </div>
                    <div className='grid grid-cols-2 gap-2'>
                        {Object.entries(urlParams).map(([key, value]) => (
                            <div key={key} className="flex flex-col gap-1">
                                <label className="text-xs text-gray-500 dark:text-gray-400 font-mono">
                                    {key}
                                </label>
                                <input
                                    type="text"
                                    value={value}
                                    onChange={(e) => setUrlParams(prev => ({
                                        ...prev,
                                        [key]: e.target.value
                                    }))}
                                    placeholder={`Enter ${key}`}
                                    className='px-3 py-2 text-sm font-mono bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md focus:outline-none focus:ring-1 focus:ring-slate-400 dark:focus:ring-slate-500 focus:border-slate-400 dark:focus:border-slate-500 transition-all'
                                />
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
        
        <div className='w-full flex flex-col mb-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-100 dark:border-gray-700 px-4 py-3 h-max'>
            <div className='w-full flex justify-between items-center gap-2 mb-3'>
                <div className='text-base font-bold text-gray-700 dark:text-gray-300 flex items-center'>
                    <span className='mr-2'>Headers</span>
                    <span className='text-xs bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-400 px-2 py-0.5 rounded-full'>{headers.length}</span>
                </div>
                <button 
                    onClick={addHeader}
                    className='flex justify-center items-center gap-1 text-sm font-mono bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 border border-gray-200 dark:border-gray-600 transition-all duration-200 rounded-md px-3 py-1.5 cursor-pointer'
                >
                    <IconPlus stroke={2} size={16} className='text-gray-600 dark:text-gray-400' />
                    <span>Add</span>
                </button>
            </div>
            
            <div className='space-y-3 pb-1'>
                {headers.map((header, index) => (
                    <div key={index} className='flex items-center gap-3 group'>
                        <input
                            type="text"
                            placeholder="Key"
                            value={header.key}
                            onChange={(e) => updateHeader(index, 'key', e.target.value)}
                            className='flex-1 px-3 py-2 text-sm font-mono bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md focus:outline-none focus:ring-1 focus:ring-slate-400 dark:focus:ring-slate-500 focus:border-slate-400 dark:focus:border-slate-500 transition-all'
                        />
                        <input
                            type="text"
                            placeholder="Value"
                            value={header.value}
                            onChange={(e) => updateHeader(index, 'value', e.target.value)}
                            className='flex-1 px-3 py-2 text-sm font-mono bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md focus:outline-none focus:ring-1 focus:ring-slate-400 dark:focus:ring-slate-500 focus:border-slate-400 dark:focus:border-slate-500 transition-all'
                        />
                        <button 
                            onClick={() => removeHeader(index)}
                            className='p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-md transition-all duration-200 opacity-50 cursor-pointer group-hover:opacity-100'
                        >
                            <IconX stroke={2} size={16} />
                        </button>
                    </div>
                ))}
            </div>
        </div>
        
        {/* Body Section - Only show for non-GET requests */}
        {method !== 'GET' && (
            <div className='w-full flex flex-col mb-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-100 dark:border-gray-700 px-4 py-3 h-max'>
                <div className='text-base font-bold text-gray-700 dark:text-gray-300 flex items-center mb-3'>
                    <IconCode size={18} className='mr-2 text-gray-600 dark:text-gray-400' />
                    <span>Request Body</span>
                </div>
                
                {isMultipart ? (
                    <div className="flex flex-col gap-2">
                        <input
                            type="file"
                            onChange={(e) => setFile(e.target.files?.[0] || null)}
                            className="w-full px-3 py-2 text-sm bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md"
                            placeholder='Select a file'
                        />
                        {file && (
                            <div className="text-sm text-gray-600 dark:text-gray-400">
                                Selected file: {file.name} ({(file.size / 1024).toFixed(2)} KB)
                            </div>
                        )}
                    </div>
                ) : (
                    <textarea
                        placeholder="Enter request body (JSON)"
                        value={body}
                        onChange={handleBodyChange}
                        className='w-full px-3 py-2 text-sm font-mono bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md focus:outline-none focus:ring-1 focus:ring-slate-400 dark:focus:ring-slate-500 focus:border-slate-400 dark:focus:border-slate-500 transition-all min-h-[120px]'
                    />
                )}
            </div>
        )}
        
        {/* Error Message */}
        {error && (
            <div className='mb-4 p-3 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 rounded-lg flex items-start'>
                <IconAlertCircle size={18} className='mr-2 flex-shrink-0 mt-0.5' />
                <div>
                    <p className='font-medium'>Error</p>
                    <p className='text-sm'>{error}</p>
                </div>
            </div>
        )}
        
        {/* Response Section */}
        {response && (
            <div className='border rounded-lg overflow-hidden border-gray-200 dark:border-gray-700 mb-2'>
                <div className='flex items-center justify-between p-3 border-b bg-gray-50 dark:bg-gray-900 border-gray-200 dark:border-gray-700'>
                    <div className='flex items-center'>
                        <IconCloudDownload size={18} className='mr-2 text-gray-600 dark:text-gray-400' />
                        <span 
                            className={`inline-block px-2 py-1 rounded text-xs font-medium ${
                                response.status < 300 ? 
                                    'bg-green-100 dark:bg-green-900/50 text-green-800 dark:text-green-400' : 
                                response.status < 400 ? 
                                    'bg-blue-100 dark:bg-blue-900/50 text-blue-800 dark:text-blue-400' : 
                                response.status < 500 ? 
                                    'bg-yellow-100 dark:bg-yellow-900/50 text-yellow-800 dark:text-yellow-400' : 
                                    'bg-red-100 dark:bg-red-900/50 text-red-800 dark:text-red-400'
                            }`}
                        >
                            {response.status} {response.statusText}
                        </span>
                    </div>
                    <div className='text-sm text-gray-500 dark:text-gray-400 flex items-center'>
                        <IconClock size={16} className='mr-1' />
                        {response.time}ms
                    </div>
                </div>
                
                <div className='p-3 overflow-auto max-h-96 bg-gray-50 dark:bg-gray-900'>
                    <pre className='text-sm font-mono whitespace-pre-wrap text-gray-800 dark:text-gray-200'>
                        {typeof response.data === 'object' 
                            ? JSON.stringify(response.data, null, 2) 
                            : response.data}
                    </pre>
                </div>
            </div>
        )}
        <div className="flex justify-end w-full">
                <button 
                    className={`w-full bg-slate-600 hover:bg-slate-500 text-sm transition-all duration-300 text-white font-mono px-3 py-2 rounded-lg cursor-pointer flex justify-center items-center disabled:opacity-50 disabled:cursor-not-allowed`}
                    onClick={sendRequest}
                    disabled={loading}
                >
                    {loading ? (
                        <>
                            <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            Sending...
                        </>
                    ) : (
                        <>
                            <IconSend stroke={2} size={18} className='mr-2' />
                            Request
                        </>
                    )}
                </button>
            </div>
    </div>
);
}