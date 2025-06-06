import React, { useState, useEffect, useRef } from 'react';

interface Message {
  message: string;
}
// extend React's keyboard event to access the isComposing flag (for example: traditional chinese)
interface MyKeyboardEvent extends React.KeyboardEvent<HTMLTextAreaElement> {
  nativeEvent: KeyboardEvent & { isComposing?: boolean };
}

const App: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState('');
  const ws = useRef<WebSocket | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const baseUrl = process.env.REACT_APP_WS_BASE_URL || 'ws://localhost:8080';
    const path = process.env.REACT_APP_WS_PATH || '/ws';
    const fullUrl = `${baseUrl}${path}`;
    // connect to backend
    ws.current = new WebSocket(fullUrl);
    // listen for messages from the server
    ws.current.onmessage = (event) => {
      console.log("event data", event.data)
      const lines = event.data.split(/\r?\n/);
      console.log("lines", lines)
      const newMessages = lines.map((line: string) => JSON.parse(line));
      setMessages((prev) => [...prev, ...newMessages]);
    };

    return () => {
      // if ws.current exists close it
      ws.current?.close();
    };
  }, []);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const sendMessage = () => {
    if (newMessage.trim() !== '' && ws.current) {
      ws.current.send(JSON.stringify({ message: newMessage }));
      setNewMessage('');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.nativeEvent.isComposing) return;
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault(); // prevent adding newline
      sendMessage();
    }
  };

  return (
    <div className="h-screen flex items-center justify-center bg-gray-100">
      <div className="w-full max-w-xl h-[90vh] bg-white shadow-lg rounded-lg flex flex-col overflow-hidden">
        {/* Header */}
        <div className="bg-blue-600 text-white text-center py-3 font-bold text-lg">
          Chat
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {messages.map((msg, index) => (
            <div key={index} className="bg-blue-100 text-gray-800 px-4 py-2 rounded-lg w-fit max-w-xs whitespace-pre-wrap break-words">
              {msg.message}
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <div className="border-t p-3 bg-gray-50">
          <div className="flex gap-2">
            <textarea
              className="flex-1 p-3 border rounded-md resize-none bg-blue-50 focus:outline-none focus:ring-2 focus:ring-blue-400 text-gray-800 h-24"
              placeholder="Type your message..."
              value={newMessage}
              onChange={(e) => setNewMessage(e.target.value)}
              onKeyDown={handleKeyDown}
            />
            <button
              className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
              onClick={sendMessage}
            >
              ➤
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default App;

