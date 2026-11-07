# jobshop manifesto

jobshop is my take on ai orchestration, inspired by hermes, gascity, and workflow engines. it is inspired by concepts borrowed largely from gastown/gascity, hermes/openclaw, and jcode/opencode.

this manifesto is completely human written. i have no issue with ai-generated content, but this readme is meant to be read by humans not agents, and so i believe that it should be written by a human.

below are details and ramblings about what jobshop is and what problems it attempts to solve. if you want technical documentation, look at the `docs` folder, but probably just point your agent at llm/README.md.

tldr; jobshop is a workflow execution framework for orchestrating the completion of tasks across multiple workable units, most notably an agent session over acp, in order to complete tasks.

## why does this exist

if you are going to use a piece of software, you should understand why the software came to exist, what problems it is meant to solve, and how it attempts to solve them.

if you attempt to use jobshop without the full wwwwwh. ai tooling is still primative and fiddly, and I don't think we are at the stage where our solutions are robust enough to truly be plug and play.

## origin

i don't like writing code. to me it is not particularly stimulating, incredibly tedious, and i have personally never enjoyed spending time on it. that said, i've written a lot of code, since usually it's a lot faster than the alternatives. since the initial introduction of llms, i had been patiently waiting for them to evolve to a point where they could be fully capable of implementation, though it took a lot longer than I would have thought...

as i've explored and used a variety of llm tooling around development and research, none had the sort of flexibility that i wanted to investigate different workflows, and this project is the result of said musings around what a good tool in my mind should provide.

## concepts, theories

IMO, the current state of llms and the tools we have around them still have technical restrictions which result in low quality output if the development is not lead by a human, or not done through burning a prohibitivly high amount of compute in iteration cycles.

it is commonly shared advice around working with llms to let the model do the work. my hypothesis is that effective ai tools are not the ones which attempt to enchance the quality of the llm and its output, but ones which address things that the llm is incapable of performing.

my logic is as follow. the LLM is likely the most expensive part of your development workflow, self-hosted or not. if you are paying this much for this product - it better be good, and it should outperform anything else that you are using by far for the tasks it is meant to do, and if it's not, then you should not be using it in the first place.

when trying to distill what an llm is actually good at, i think that they are incredibly effective translation machines. they are able to translate data from one language to another, while maintaining the underlying "semantics" reflected by the input. just as a good translation machine translates between english and chinese while maintaining the underlying meaning of the phrases, not just the direct translation of the symbols, LLMs do the same sort of translation, just from languages like "question->answer", or "specification -> implementation", and vice versa.

what i think makes LLMs useful is that their concept of what a "language" can be, is rooted in the weights used for its translations, which means that any concept that can be represented in the weights is a valid language, both for input and output. so if two concepts are included in the weights, then an arbitrary datum can be translated as something that is originally one concept to the other. this is what creates the ?illusion? that llms can "write everything", because in many ways, with enough tuned weights, there are so many concepts encoded that they basically can.

from this, what i believe is that the shortcoming of the LLM is when it is not given a task that can clearly be represented as convert data in input language to data in output language. what this also means is that when an stronger model is able to accomplish a task, and a smaller is not, it is because the small model was unable to achieve one of the steps, figuring out the input/output language, or having the weights necessary to having a good translation between the two for the data.

i think this lines up very well with the llm tooling that is useful, and the llm tooling that is not. things that coding harnesses, agent harnesses, etc do:

"tools" and "tool calls" are a great example of this, say a tool that calls a shell script in a linux machine. runtimes provide tools to models because text alone is not able to say, write a file, or execute a binary. these are things that require a mappings of input and output via side effects to be translated in chains. tools address a task that LLMs are simply unable to do, and therefore unlock a great amount of their value.

"AGENTS.md", or the concept of injected context into every session. providing injected context through many sessions fixes two big problems. first, they provide a conceptual bridge to help models map input to the correct input and output languages, by providing context relevant to what it knows the user will end up doing, which can allow a model that doesnt have the mapping of the question without context to be able to answer the question. second, they reduce the overall context size of each session, as the user spawns new sessions and sub sessions with fresh context, and performs compaction with routines that prefer to keep initial instructions, which not only helps reduce mismappings as a result of unrelated content, but also reduce costs, which is important, as less cost = more runs = better output.

"memories", "self learning", "self improvement loops" which is really just, encoding into context translation mappings that you commonly do, and therefore are more likely the correct ones.

running agent runtimes in ci, such as code review agents, are really just providing a task itself for the llm to complete. the useful thing here is the trigger, the idea that you are transforming a side effect (which the llm cant handle), into a translation task (code diffs + rules -> code review), and so now the llm is unlocked.

"todo lists", "goal", "hill-climbing", "rounds", "multi-agent", allow agents to explore an indirect path, instead of going from language a->c directly, it allows for the path to be from a->b->c, which unlocks models that do not have weights for a->c to still be able to do the translation. separating tasks in swarms is so something like unlocking the round |a->b,a->c,a->d|b+c+d->e|, instead of a->e

in contrast, tools and solutions that don't provide these things have been found to not be so useful.

a very clear example is the self-review-in-same-session footgun. here, an agent is instructed to translate specification->implementation, and then a translation of implementation->specification. a fresh session allows new translation paths to be discovered because it may not necessarily perform the same translation paths of the initial implementation and therefore find new things, and many use another model, which is a completely different mapping, and so different to review outcomes will be discovered.

another good example is providing agents with "titles". i suspect this is because title prompts do not meaningfully change the quality of the weights themselves. the weights are already trained (it's the most expensive part), and non specific titles (expert, junior), do not meaningfully change the best trained mapping that the llm is already tuned to try to deliver.

the same applies to prompts that attempt to "simulate a stronger model". mappings that do not exist cannot be created, and the highest quality mappings given the inputs are always what the model attempts to output.

## a useful tool

from the theory above, the idea for jobshop is that it should only provide features which are useful under this metric. a secondary usefulness metric more obvious is the token and compute efficiency. less cycles = more work. sadly i don't have infinite compute.

that is not to say that all features will be useful, just that every useful feature should be quantifiable as something that directly addresses a weakness of the llm, allowing a set of weights to be able to achieve something that it could not have previously done as a single input->output step, whether that because of non-comprehensive weights, or simply just not a correctly parametized translation task.

finally, we arrive at the feature set of jobshop, none of which are original or unique.

1. the ability to create and control persistent compacting sessions, with ability to inject per role. there's no need to reinvent the tool call and harness layer and the advantages that come with that, and so we wrap arbitrary clients with acp. clearly there is value of llms able to have thier output write files and run commands.
2. track work over time with annotatable and linkable tickets.  an internal ticketing system allows the translation to go from a->z instead from a->b...->z, keeping the context and inputs focused per translation
3. ingest realtime hooks and reformat them to be delivered as llm workable tasks
4. transform and route the communication between tasks with structured rules to ensure that tasks can be given to agents with the correct context, and tasks are formatted in correctly translatable formats
5. provide mechanisms to communicate with external systems, be it a human, another agent, an issue tracker, etc, so that original inputs can be fed in

good luck with jobshop!
