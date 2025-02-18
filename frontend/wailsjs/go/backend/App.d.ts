// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {backend} from '../models';

export function GetCaptureStats():Promise<backend.CaptureStats>;

export function GetInterfaces():Promise<Array<string>>;

export function GetPackets():Promise<Array<backend.Packet>>;

export function GetPingResults(arg1:string):Promise<Array<backend.PingResult>>;

export function Greet(arg1:string):Promise<string>;

export function StartCapture(arg1:string,arg2:string):Promise<void>;

export function StartPing(arg1:string,arg2:number):Promise<void>;

export function StopCapture():Promise<void>;

export function StopPing(arg1:string):Promise<void>;

export function TestTCPConnection(arg1:string,arg2:number,arg3:number):Promise<backend.TCPResult>;
