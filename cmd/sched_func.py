import os, sys
sys.path.insert(0, './pkg/workload/schedproto')
import time 
from concurrent import futures
import logging
import grpc
import sched_pb2
import sched_pb2_grpc
from dataclasses import dataclass
import random 

# Configure the logging system
logging.basicConfig(format='%(asctime)s - %(levelname)s - %(message)s', 
                    datefmt='%Y/%m/%d %H:%M:%S', level=logging.INFO)


IDLE='idle'
RUNNING='running'
TotalGPU=40
class Empty(object): 
    pass 

@dataclass 
class Job: 
    name: str
    batchsize: int 
    deadline: int 
    iterations: int 
    prevReplica: int 
    

def print_red_text(text):
    RED = "\033[91m"
    RESET = "\033[0m"
    print(RED + text + RESET, flush=True)
          

DEBUG = True 
class Executor(sched_pb2_grpc.Executor):
    def __init__(self, ): 
        super().__init__()
        self.sched_interval = 10 # seconds 
        self.llama_cnt = 0 
        
    def Execute(self, request, context, **kwargs):
        return sched_pb2.SchedReply(replica=1, schedOverhead=1)

    def ExecuteStream(self, request_iterator, context, **kwargs):
        # print("starting running", time.time())
        start = time.time()
        job_infos = list() 
        name_keys = list() 
        remaining_gpus = 0 
        sched_alg = None 
        for request in request_iterator:
            job_infos.append(Job(name=request.invocationName, batchsize=request.batchsize, \
                            deadline=request.deadline, iterations=request.iterations, prevReplica=request.prevReplica))
            name_keys.append(request.invocationName)
            remaining_gpus = request.availableGPU
            sched_alg = request.schedAlg
        if remaining_gpus > 0: 
            logging.info(f"sched_alg {sched_alg}, remaining_gpus {remaining_gpus}")
        if sched_alg in ['elastic_flow', 'infless', 'elastic']: 
            num_replicas = {name:0 for name in name_keys}
            name = name_keys[0]
            if remaining_gpus >= 32: 
                num_replicas[name] = 32
            # elif remaining_gpus >= 16: 
            #     num_replicas[name] = 16
            # elif remaining_gpus >= 8: 
            #     num_replicas[name] = 8 
            # elif remaining_gpus >= 4: 
            #     num_replicas[name] = 4
                
        ret_replicas = [int(num_replicas[name]) for name in name_keys]
        if sum(num_replicas.values()) > 0: 
            print(name_keys, flush=True)
            print(ret_replicas, flush=True)
        response = sched_pb2.SchedReply(invocationName=name_keys, replica=ret_replicas, schedOverhead=int(time.time()-start))
        return response 

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    sched_pb2_grpc.add_ExecutorServicer_to_server(Executor(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    print('starting sever ...', flush=True)
    logging.basicConfig()
    serve()
